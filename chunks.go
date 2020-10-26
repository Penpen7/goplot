package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
)

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
	// fmt.Println("size", size)

	// seek 4byte
	n := make([]byte, FOOTERSIZE)
	_, err := file.Read(n)
	if err == io.EOF {
		fmt.Println("End")
	}
	return bytes.NewBuffer(m)
}
func writeFieldData(g []float32, config simulationConfig, fname string, wg *sync.WaitGroup) {
	fout, err := os.Create(fname)
	if err != nil {
		panic(err)
	}

	index := 0
	for z := int32(1); z <= config.OutputMeshNumber[2]; z++ {
		for y := int32(1); y <= config.OutputMeshNumber[1]; y++ {
			for x := int32(1); x <= config.OutputMeshNumber[0]; x++ {
				fout.WriteString(fmt.Sprintln(x, y, z, g[index]))
				index++
			}
			fout.WriteString("\n")
		}
		fout.WriteString("\n")
	}
	fout.Close()
	wg.Done()
}

func loadWriteFieldData(file *os.File, config simulationConfig, fileID int, wg *sync.WaitGroup) {
	var totalMeshNumber = config.MeshNumber[0] * config.MeshNumber[1] * config.MeshNumber[2]
	//Ex
	title := [...]string{"Ex", "Ey", "Ez", "Bx", "By", "Bz", "Jx", "Jy", "Jz"}
	for _, v := range title {
		fmt.Printf("loading... %s\n", v)
		g := []float32{}
		g = make([]float32, totalMeshNumber)
		binary.Read(readNextChunk(file), binary.LittleEndian, &g)

		wg.Add(1)
		go writeFieldData(g[:], config, fmt.Sprintf("%s_%04d.txt", v, fileID), wg)
	}
}

func loadSnap(file *os.File, config simulationConfig, fileID int) {
	var time float32
	wg := &sync.WaitGroup{}

	binary.Read(readNextChunk(file), binary.LittleEndian, &time)
	fmt.Println("time", time)

	loadWriteFieldData(file, config, fileID, wg)
	title_particle := [...]string{"Ion_Density", "Ion_Energy", "Ion_EnergyFlux_x", "Ion_EnergyFlux_y"}

	for ionID := int32(1); ionID <= config.IonNumber; ionID++ {
		for _, v := range title_particle {
			fmt.Printf("read %s\n", v)
			g := []float32{}
			g = make([]float32, config.TotalMeshNumber)
			binary.Read(readNextChunk(file), binary.LittleEndian, &g)
			// writeData(g[:], v)
		}
	}

	title_particle_Electron := [...]string{"Electron_Density", "Electron_Energy", "Electron_EnergyFlux_x", "Electron_EnergyFlux_y"}
	for ElectronID := int32(1); ElectronID <= config.ElectronNumber; ElectronID++ {
		for _, v := range title_particle_Electron {
			g := []float32{}
			g = make([]float32, config.TotalMeshNumber)
			fmt.Println(v)
			binary.Read(readNextChunk(file), binary.LittleEndian, &g)
			// writeData(g[:], v)
		}
	}
	for i := int32(1); i <= config.TotalParticleSpecies; i++ {
		var dltmomentum float32
		momentumvsmomentum := []float32{}
		momentumvsmomentum = make([]float32, config.MomentumMeshNumber*config.MomentumMeshNumber)
		binary.Read(readNextChunk(file), binary.LittleEndian, &dltmomentum)
		momentum_title := [...]string{"pxpy", "pypz", "pzpx"}
		for _, v := range momentum_title {
			fmt.Println(v)
			binary.Read(readNextChunk(file), binary.LittleEndian, &momentumvsmomentum)
		}
		position_title := [...]string{"xpx", "xpy", "xpz", "ypx", "ypy", "ypz"}
		for _, v := range position_title {
			fmt.Println(v)
			readNextChunk(file)
		}

		var dltvelocity float32
		velocityvsvelocity := []float32{}
		velocityvsvelocity = make([]float32, config.MomentumMeshNumber*config.MomentumMeshNumber)
		binary.Read(readNextChunk(file), binary.LittleEndian, &dltvelocity)
		velocity_title := [...]string{"pxpy", "pypz", "pzpx"}
		for _, v := range velocity_title {
			fmt.Println(v)
			binary.Read(readNextChunk(file), binary.LittleEndian, &velocityvsvelocity)
		}
		position_velocity_title := [...]string{"xpx", "xpy", "xpz", "ypx", "ypy", "ypz"}
		for _, v := range position_velocity_title {
			fmt.Println(v)
			readNextChunk(file)
		}
	}
	for i := int32(1); i <= config.TotalParticleSpecies; i++ {
		fmt.Println("energy", i)
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
		fmt.Println("energy", i)
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
	wg.Wait()
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
	TotalMeshNumber            int32
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

func loadSetting() simulationConfig {
	var config simulationConfig
	file, _ := os.Open("gfin.dat")
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

	config.TotalMeshNumber = config.MeshNumber[0] * config.MeshNumber[1] * config.MeshNumber[2]
	return config
}
func showConfig(config simulationConfig) {
	fmt.Printf("%+v\n", config)
	// fmt.Printf("%-10s %v\n", "バージョン", config.Version)
	// fmt.Printf("%-10s %v\n", "並列数", config.ParallelNumber)
	// fmt.Printf("%-10s %v\n", "次元", config.Dimension)
}

func main() {
	file, err := os.Open("snap0001.dat")
	if err != nil {
		fmt.Printf("error!")
		os.Exit(-1)
	}
	config := loadSetting()
	showConfig(config)
	for i := 0; i < 3; i++ {
		loadSnap(file, config, i)
	}
}
