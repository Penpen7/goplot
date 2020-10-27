package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const outputDirectory = "biny_data2"

func readNextChunk(file *os.File) *bytes.Buffer {
	const HEADERSIZE = 4
	const FOOTERSIZE = 4
	l := make([]byte, HEADERSIZE)

	// seek 4byte
	file.Read(l)
	var size int32
	binary.Read(bytes.NewBuffer(l), binary.LittleEndian, &size)

	// seek size byte
	m := make([]byte, size)
	binary.Read(file, binary.LittleEndian, &m)

	// seek 4byte
	n := make([]byte, FOOTERSIZE)
	_, err := file.Read(n)
	if err == io.EOF {
		fmt.Println("ファイルの終端に達しました")
		os.Exit(0)
	} else if err != nil {
		panic(err)
	}
	return bytes.NewBuffer(m)
}

func writeFieldData(g []float32, config simulationConfig, fname string, wg *sync.WaitGroup) {
	fout, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(fout)

	index := 0
	for z := int32(1); z <= config.OutputMeshNumber[2]; z++ {
		for y := int32(1); y <= config.OutputMeshNumber[1]; y++ {
			for x := int32(1); x <= config.OutputMeshNumber[0]; x++ {
				writer.WriteString(fmt.Sprintln(x, y, z, g[index]))
				index++
			}
			writer.WriteString("\n")
		}
		writer.WriteString("\n")
	}
	writer.Flush()
	fout.Close()
	wg.Done()
}

func loadWriteFieldData(file *os.File, config simulationConfig, fileID int, wg *sync.WaitGroup) {
	title := [...]string{"Ex", "Ey", "Ez", "Bx", "By", "Bz", "Jx", "Jy", "Jz"}
	for _, v := range title {
		fmt.Printf("loading... %s\n", v)
		g := []float32{}
		g = make([]float32, config.TotalOutputMeshNumber)
		binary.Read(readNextChunk(file), binary.LittleEndian, &g)

		wg.Add(1)
		go writeFieldData(g[:], config, fmt.Sprintf("%s/%s_%04d.txt", outputDirectory, v, fileID), wg)
	}
}

func loadWriteParticleMeshData(file *os.File, config simulationConfig, fileID int, wg *sync.WaitGroup) {

	title_particle := [...]string{"Ion_Density", "Ion_Energy", "Ion_EnergyFlux_x", "Ion_EnergyFlux_y"}
	title_particle_Electron := [...]string{"Electron_Density", "Electron_Energy", "Electron_EnergyFlux_x", "Electron_EnergyFlux_y"}

	for ionID := int32(1); ionID <= config.IonNumber; ionID++ {
		for _, v := range title_particle {
			fmt.Printf("loading... %s\n", v)
			g := []float32{}
			g = make([]float32, config.TotalOutputMeshNumber)
			binary.Read(readNextChunk(file), binary.LittleEndian, &g)
			wg.Add(1)
			go writeFieldData(g[:], config, fmt.Sprintf("%s/%s%04d_is=%02d.txt", outputDirectory, v, fileID, ionID), wg)
		}
	}

	for ElectronID := int32(1); ElectronID <= config.ElectronNumber; ElectronID++ {
		for _, v := range title_particle_Electron {
			g := []float32{}
			g = make([]float32, config.TotalOutputMeshNumber)
			fmt.Printf("loading... %s\n", v)
			binary.Read(readNextChunk(file), binary.LittleEndian, &g)
			wg.Add(1)
			go writeFieldData(g[:], config, fmt.Sprintf("%s/%s%04d_is=%02d.txt", outputDirectory, v, fileID, ElectronID), wg)
		}
	}
}
func loadWritePhaseSpace(file *os.File, config simulationConfig, fileID int, wg *sync.WaitGroup) {
	momentum_title := [...]string{"pxpy", "pypz", "pzpx"}
	position_title := [...]string{"xpx", "xpy", "xpz", "ypx", "ypy", "ypz"}
	velocity_title := [...]string{"pxpy", "pypz", "pzpx"}
	position_velocity_title := [...]string{"xpx", "xpy", "xpz", "ypx", "ypy", "ypz"}
	for i := int32(1); i <= config.TotalParticleSpecies; i++ {
		var dltmomentum float32
		momentumvsmomentum := []float32{}
		momentumvsmomentum = make([]float32, config.MomentumMeshNumber*config.MomentumMeshNumber)
		binary.Read(readNextChunk(file), binary.LittleEndian, &dltmomentum)
		// TODO:後で書き込みを実装
		// for _, v := range momentum_title {
		for i := int32(0); i < int32(len(momentum_title)); i++ {
			binary.Read(readNextChunk(file), binary.LittleEndian, &momentumvsmomentum)
		}

		// TODO:後で書き込みを実装
		// for _, v := range position_title {
		for i := int32(0); i < int32(len(position_title)); i++ {
			readNextChunk(file)
		}

		var dltvelocity float32
		velocityvsvelocity := []float32{}
		velocityvsvelocity = make([]float32, config.MomentumMeshNumber*config.MomentumMeshNumber)
		binary.Read(readNextChunk(file), binary.LittleEndian, &dltvelocity)
		// TODO:後で書き込みを実装
		// for _, v := range velocity_title {
		for i := int32(0); i < int32(len(velocity_title)); i++ {
			binary.Read(readNextChunk(file), binary.LittleEndian, &velocityvsvelocity)
		}
		// TODO:後で書き込みを実装
		// for _, v := range position_velocity_title {
		for i := int32(0); i < int32(len(position_velocity_title)); i++ {
			readNextChunk(file)
		}
	}
}
func loadWriteEnergyDistribution(file *os.File, config simulationConfig, fileID int, wg *sync.WaitGroup) {
	for i := int32(1); i <= config.TotalParticleSpecies; i++ {
		var averageChargeRate, averageEnergy, dltEnergy, Eimaxt float32
		binary.Read(readNextChunk(file), binary.LittleEndian, &averageChargeRate)
		binary.Read(readNextChunk(file), binary.LittleEndian, &averageEnergy)
		binary.Read(readNextChunk(file), binary.LittleEndian, &dltEnergy)
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
		binary.Read(readNextChunk(file), binary.LittleEndian, &Eimaxt)
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
	}
	for i := int32(1); i <= config.TotalParticleSpecies; i++ {
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
		readNextChunk(file)
	}
}

func loadSnap(file *os.File, config simulationConfig, fileID int) {
	var simulationTime float32
	wg := &sync.WaitGroup{}
	start := time.Now()

	binary.Read(readNextChunk(file), binary.LittleEndian, &simulationTime)
	fmt.Println("読み込んでいるシミュレーション上の規格化時間:", simulationTime)

	loadWriteFieldData(file, config, fileID, wg)
	loadWriteParticleMeshData(file, config, fileID, wg)
	loadWritePhaseSpace(file, config, fileID, wg)
	loadWriteEnergyDistribution(file, config, fileID, wg)
	fmt.Println("書き込み中...")
	wg.Wait()

	end := time.Now()
	fmt.Println("経過時間:", end.Sub(start))
	fmt.Println("")
}

type simulationConfig struct {
	Version                    string
	ParallelNumber             int32
	Dimension                  int32
	VelocityLight              float64
	DeltTime                   float64
	DeltX                      [3]float64
	SystemL                    [3]float64
	AverageDensity             float64
	MeshNumber                 [3]int32
	FildBoundaryCondition      int32
	TotalParticleNumber        int32
	TotalParticleSpecies       int32
	TotalOutputMeshNumber      int32
	IonNumber                  int32
	ElectronNumber             int32
	Loadtype                   []int32
	ClusterOption              bool
	ClusterNumber              int32
	CollisionOption            bool
	Ncol                       int32
	UsedIonize                 bool
	UsedFieldIonize            bool
	UsedCollisionalIonize      bool
	UsedHeneutralCollision     bool
	UsedIonizeFieldLoss        bool
	UsedFileIonizeADKmodel     bool
	UsedIonizeFieldKeldysh     bool
	UsedLLDumpingOption        bool
	UsedLocalSolver            bool
	RealLx                     float64
	intSnap                    int32
	OutputMeshNumber           [3]int32
	MomentumMeshNumber         int32
	SpaceMeshNumberForMomentum [2]int32
}

func loadSetting() (simulationConfig, error) {
	var config simulationConfig
	file, err := os.Open("gfin.dat")
	if err != nil {
		return config, err
	}
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.Version)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.ParallelNumber)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.Dimension)

	var buf = readNextChunk(file)
	binary.Read(buf, binary.LittleEndian, &config.VelocityLight)
	binary.Read(buf, binary.LittleEndian, &config.DeltTime)
	binary.Read(buf, binary.LittleEndian, &config.DeltX)

	binary.Read(readNextChunk(file), binary.LittleEndian, &config.SystemL)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.AverageDensity)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.MeshNumber)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.FildBoundaryCondition)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.TotalParticleNumber)

	buf = readNextChunk(file)
	binary.Read(buf, binary.LittleEndian, &config.TotalParticleSpecies)
	binary.Read(buf, binary.LittleEndian, &config.IonNumber)
	binary.Read(buf, binary.LittleEndian, &config.ElectronNumber)

	readNextChunk(file)

	binary.Read(readNextChunk(file), binary.LittleEndian, &config.ClusterOption)
	if config.ClusterOption {
		binary.Read(readNextChunk(file), binary.LittleEndian, &config.ClusterNumber)
	}
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.ClusterOption)
	if config.UsedCollisionalIonize {
		binary.Read(readNextChunk(file), binary.LittleEndian, &config.ClusterNumber)
	}
	readNextChunk(file)
	readNextChunk(file)
	readNextChunk(file)
	readNextChunk(file)
	readNextChunk(file)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.OutputMeshNumber)

	config.TotalOutputMeshNumber = config.OutputMeshNumber[0] * config.OutputMeshNumber[1] * config.OutputMeshNumber[2]
	return config, nil
}

// 設定を表示する
func showConfig(config simulationConfig) {
	fmt.Printf("%+v\n", config)
}

func main() {
	// outputDirectoryがあるか確認。なければディレクトリを作成する
	if _, err := os.Stat(outputDirectory); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(outputDirectory, 0777); err != nil {
				fmt.Println("biny_dataが作れませんでした")
				fmt.Println(err)
				os.Exit(-1)
			}
		}
	}

	// snapのバイナリを開く(とりあえずここではsnap0001.dat)
	file, err := os.Open("snap0001.dat")
	if err != nil {
		fmt.Println("snap0001.datが読み込めません")
		fmt.Println(err)
		os.Exit(-1)
	}

	// gfin.datを開き、シミュレーション設定を読み込む。
	config, err := loadSetting()
	if err != nil {
		fmt.Println("gfin.datが読み込めません")
		fmt.Println(err)
		os.Exit(-1)
	}

	// 設定を表示する
	showConfig(config)

	// snapを終端に達するまで読み込む。
	for i := 0; ; i++ {
		loadSnap(file, config, i)
	}
}
