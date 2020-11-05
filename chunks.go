package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const outputDirectoryAuthority = 0777
const plotConfigFileName = "plot.json"

var electricFieldNormalizeConstant float32
var magneticFieldNormalizeConstant float32

type subart struct {
	Name   string
	Plot   bool
	Center string
}
type art struct {
	OutputASCIIDirectory string
	OutputVTKDirectory   string
	Field                []subart
	Particle             []subart
}

var plotConfig art

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
	FormFactor                 int32
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
	IonStep                    int32
	IntSnap                    int32
}

func writeFieldData(g [][][]float32, mode string, fname string, wg *sync.WaitGroup) {
	fout, err := os.Create(fname)
	defer fout.Close()
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(fout)

	xsize := len(g)
	ysize := len(g[0])
	zsize := len(g[0][0])

	switch mode {
	case "xyz":
		for x := 0; x < xsize; x++ {
			for y := 0; y < ysize; y++ {
				for z := 0; z < zsize; z++ {
					writer.WriteString(fmt.Sprintln(x, y, z, g[x][y][z]))
				}
				writer.WriteString(fmt.Sprintln(""))
			}
			writer.WriteString(fmt.Sprintln(""))
		}
		break
	case "xy":
		for x := 0; x < xsize; x++ {
			for y := 0; y < ysize; y++ {
				writer.WriteString(fmt.Sprintln(x, y, g[x][y][zsize/2]))
			}
			writer.WriteString(fmt.Sprintln(""))
		}
		break
	case "yz":
		for y := 0; y < ysize; y++ {
			for z := 0; z < zsize; z++ {
				writer.WriteString(fmt.Sprintln(y, z, g[xsize/2][y][z]))
			}
			writer.WriteString(fmt.Sprintln(""))
		}
		break
	case "zx":
		for z := 0; z < zsize; z++ {
			for x := 0; x < xsize; x++ {
				writer.WriteString(fmt.Sprintln(z, x, g[x][ysize/2][z]))
			}
			writer.WriteString(fmt.Sprintln(""))
		}
		break
	case "x":
		for x := 0; x < xsize; x++ {
			writer.WriteString(fmt.Sprintln(x, g[x][ysize/2][zsize/2]))
		}
		break
	case "y":
		for y := 0; y < ysize; y++ {
			writer.WriteString(fmt.Sprintln(y, g[xsize/2][y][zsize/2]))
		}
		break
	case "z":
		for z := 0; z < zsize; z++ {
			writer.WriteString(fmt.Sprintln(z, g[xsize/2][ysize/2][z]))
		}
		break
	case "zxaverage":
		for z := 0; z < zsize; z++ {
			for x := 0; x < zsize; x++ {
				sum := float32(0)
				for y := 0; y < ysize; y++ {
					sum += g[x][y][z]
				}
				sum /= float32(ysize)
				writer.WriteString(fmt.Sprintln(z, x, sum))
			}
			writer.WriteString("\n")
		}
		break
	case "xaverage":
		average := make([][]float32, xsize)
		for x := 0; x < xsize; x++ {
			average[x] = make([]float32, ysize)
		}
		for x := 0; x < xsize; x++ {
			for y := 1; y < ysize; y++ {
				average[x][y] += average[x][y-1] + g[x][y][zsize/2]
			}
			var averagieze func(int) float32 = func(n int) float32 {
				return (average[x][n*ysize/8] - average[x][(n-1)*ysize/8]) / (float32(ysize) / 8)
			}
			writer.WriteString(fmt.Sprintln(x, averagieze(1), averagieze(2), averagieze(3), averagieze(4),
				averagieze(5), averagieze(6), averagieze(7)))
		}
		break
	default:
		fmt.Println("Warning:invalid mode:", mode)
	}
	writer.Flush()
	fout.Close()
	wg.Done()
}

func slice1Dto3D(slice1D []float32, xsize int32, ysize int32, zsize int32, normalizeConstant float32) [][][]float32 {
	// 3次元スライス作成
	var slice3D [][][]float32
	slice3D = make([][][]float32, xsize)
	for x := int32(0); x < xsize; x++ {
		slice3D[x] = make([][]float32, ysize)
		for y := int32(0); y < ysize; y++ {
			slice3D[x][y] = make([]float32, zsize)
		}
	}

	// 1次元配列から、3次元配列に割り当てる。
	index2D := 0
	for y := int32(0); y < ysize; y++ {
		for z := int32(0); z < zsize; z++ {
			for x := int32(0); x < xsize; x++ {
				slice3D[x][y][z] = slice1D[index2D] * normalizeConstant
				index2D++
			}
		}
	}

	return slice3D
}
func readNextChunk(file *os.File) *bytes.Buffer {
	// TODO:読み込みを並行化させた方がいい
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
		return nil
	} else if err != nil {
		panic(err)
	}
	return bytes.NewBuffer(m)
}
func writeFieldVTK(g [][][]float32, fname string, arrayName string, config simulationConfig, wg *sync.WaitGroup) {
	fout, err := os.Create(fname)
	defer fout.Close()
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(fout)
	writer.WriteString("<?xml version=\"1.0\"?>\n")
	writer.WriteString("<VTKFile type=\"ImageData\" byte_order=\"LittleEndian\">")
	writer.WriteString(fmt.Sprintf("<ImageData WholeExtent=\"0 %d 0 %d 0 %d\" Origin=\"0 0 0\" Spacing=\"1.0 1.0 1.0\">", config.OutputMeshNumber[0]-1, config.OutputMeshNumber[1]-1, config.OutputMeshNumber[2]-1))
	writer.WriteString(fmt.Sprintf("<Piece Extent=\"0 %d 0 %d 0 %d\">", config.OutputMeshNumber[0]-1, config.OutputMeshNumber[1]-1, config.OutputMeshNumber[2]-1))
	writer.WriteString(fmt.Sprintf("<PointData Scalars=\"%s\">", arrayName))
	writer.WriteString(fmt.Sprintf("<DataArray Name=\"%s\" type=\"Float32\" format=\"binary\">", arrayName))

	var dataSizeInByte int32
	dataSizeInByte = config.TotalOutputMeshNumber * 4
	var dataSizeInByte2Byte []byte
	dataSizeBuffer := bytes.NewBuffer(dataSizeInByte2Byte)
	binary.Write(dataSizeBuffer, binary.LittleEndian, dataSizeInByte)
	writer.WriteString(base64.StdEncoding.EncodeToString(dataSizeBuffer.Bytes()))

	var a []byte
	buf := bytes.NewBuffer(a)
	xsize := len(g)
	ysize := len(g[0])
	zsize := len(g[0][0])
	for x := 0; x < xsize; x++ {
		for y := 0; y < ysize; y++ {
			for z := 0; z < zsize; z++ {
				binary.Write(buf, binary.LittleEndian, g[x][y][z])
			}
		}
	}
	writer.WriteString(base64.StdEncoding.EncodeToString(buf.Bytes()))

	writer.WriteString("</DataArray></PointData></Piece></ImageData></VTKFile>")
	writer.Flush()
	fout.Close()
	wg.Done()
}
func loadWriteFieldData(file *os.File, config simulationConfig, fileID int, wg *sync.WaitGroup) {
	title := [...]string{"Ex", "Ey", "Ez", "Bx", "By", "Bz", "Jx", "Jy", "Jz"}
	normalizeConst := [...]float32{electricFieldNormalizeConstant, electricFieldNormalizeConstant, electricFieldNormalizeConstant,
		magneticFieldNormalizeConstant, magneticFieldNormalizeConstant, magneticFieldNormalizeConstant,
		1.0, 1.0, 1.0}
	for i, v := range title {
		fmt.Printf("\r\033[K loading... %s", v)
		g := []float32{}
		g = make([]float32, config.TotalOutputMeshNumber)
		nextchunk := readNextChunk(file)
		binary.Read(nextchunk, binary.LittleEndian, &g)
		buf := slice1Dto3D(g, config.OutputMeshNumber[0], config.OutputMeshNumber[1], config.OutputMeshNumber[2], normalizeConst[i])

		for _, vconfig := range plotConfig.Field {
			if vconfig.Name == v {
				if !vconfig.Plot {
					break
				}
				for _, vcenter := range strings.Split(vconfig.Center, " ") {
					wg.Add(1)
					if vcenter == "vtk" {
						go writeFieldVTK(buf, fmt.Sprintf("%s/%s%04d.vti", plotConfig.OutputVTKDirectory, v, fileID), v, config, wg)
					} else {
						go writeFieldData(buf, vcenter, fmt.Sprintf("%s/%s_%s_%04d.txt", plotConfig.OutputASCIIDirectory, v, vcenter, fileID), wg)
					}
				}
			}
		}
	}
}

func loadWriteParticleMeshData(file *os.File, config simulationConfig, fileID int, wg *sync.WaitGroup) {

	title_particle := [...]string{"Ion_Density", "Ion_Energy", "Ion_EnergyFlux_x", "Ion_EnergyFlux_y"}
	title_particle_Electron := [...]string{"Electron_Density", "Electron_Energy", "Electron_EnergyFlux_x", "Electron_EnergyFlux_y"}

	for ionID := int32(1); ionID <= config.IonNumber; ionID++ {
		for _, v := range title_particle {
			fmt.Printf("\r\033[K loading... %s", v)
			g := []float32{}
			g = make([]float32, config.TotalOutputMeshNumber)
			binary.Read(readNextChunk(file), binary.LittleEndian, &g)
			buf := slice1Dto3D(g, config.OutputMeshNumber[0], config.OutputMeshNumber[1], config.OutputMeshNumber[2], 1.0)

			for _, vconfig := range plotConfig.Particle {
				if vconfig.Name == v && vconfig.Plot {
					for _, vplot := range strings.Split(vconfig.Center, " ") {
						wg.Add(1)
						if vplot == "vtk" {
							go writeFieldVTK(buf, fmt.Sprintf("%s/%s%04d_is=%02d.vti", plotConfig.OutputVTKDirectory, v, fileID, ionID), v, config, wg)
						} else {
							go writeFieldData(buf, vplot, fmt.Sprintf("%s/%s_%s_%04d_is=%02d.txt", plotConfig.OutputASCIIDirectory, v, vplot, fileID, ionID), wg)
						}
					}
				}
			}
		}
	}

	for ElectronID := config.IonNumber + 1; ElectronID <= config.TotalParticleSpecies; ElectronID++ {
		for _, v := range title_particle_Electron {
			g := []float32{}
			g = make([]float32, config.TotalOutputMeshNumber)
			fmt.Printf("\r\033[K loading... %s", v)
			binary.Read(readNextChunk(file), binary.LittleEndian, &g)
			buf := slice1Dto3D(g, config.OutputMeshNumber[0], config.OutputMeshNumber[1], config.OutputMeshNumber[2], 1.0)
			for _, vconfig := range plotConfig.Particle {
				if vconfig.Name == v && vconfig.Plot {
					for _, vplot := range strings.Split(vconfig.Center, " ") {
						wg.Add(1)
						if vplot == "vtk" {
							go writeFieldVTK(buf, fmt.Sprintf("%s/%s%04d_is=%02d.vti", plotConfig.OutputVTKDirectory, v, fileID, ElectronID), v, config, wg)
						} else {
							go writeFieldData(buf, vplot, fmt.Sprintf("%s/%s_%s_%04d_is=%02d.txt", plotConfig.OutputASCIIDirectory, v, vplot, fileID, ElectronID), wg)
						}
					}
				}
			}
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
	// TODO:後で書き込みを実装
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

func loadSnap(file *os.File, config simulationConfig, fileID int) bool {
	var simulationTime float32
	wg := &sync.WaitGroup{}
	start := time.Now()

	binaryReader := readNextChunk(file)
	if binaryReader == nil {
		return true
	}
	binary.Read(binaryReader, binary.LittleEndian, &simulationTime)
	fmt.Println("")
	fmt.Println("読み込んでいるシミュレーション上の規格化時間:", simulationTime)

	loadWriteFieldData(file, config, fileID, wg)
	loadWriteParticleMeshData(file, config, fileID, wg)
	loadWritePhaseSpace(file, config, fileID, wg)
	loadWriteEnergyDistribution(file, config, fileID, wg)
	fmt.Printf("\r\033[K書き込み中...")
	wg.Wait()
	fmt.Printf("\r\033[K書き込み完了\n")
	end := time.Now()
	fmt.Println("経過時間:", end.Sub(start))
	fmt.Println("")
	return false
}
func calculateNormalizeConstant(sc simulationConfig) {
	lightSpeed := 2.99792458e+10   //c_r
	electronMass := 9.10938356e-28 //rme_r
	electricUnit := 4.8032e-10     //e_r
	// electronVoltToJoule := 1.602e-12 //eV_J_r
	normalizedDeltaX := sc.RealLx / sc.SystemL[0]
	normalizedPlasmaFrequency := lightSpeed / sc.VelocityLight / normalizedDeltaX
	normalizedNumberDensity := math.Pow(normalizedPlasmaFrequency, 2) * electronMass / (4.0 * math.Pi * math.Pow(electricUnit, 2))
	electricFieldNormalizeConstant = float32(4.0 * math.Pi * normalizedNumberDensity * electricUnit * normalizedDeltaX * 1e+4 * 3.0)
	magneticFieldNormalizeConstant = electricFieldNormalizeConstant / float32(lightSpeed*1e-2)
	// fmt.Println(normalizedDeltaX, normalizedPlasmaFrequency, normalizedNumberDensity, electricFieldNormalizeConstant, magneticFieldNormalizeConstant)
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

	config.Loadtype = make([]int32, config.TotalParticleSpecies)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.Loadtype)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.ClusterOption)
	if config.ClusterOption {
		binary.Read(readNextChunk(file), binary.LittleEndian, &config.ClusterNumber)
	}
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.CollisionOption)
	if config.CollisionOption {
		binary.Read(readNextChunk(file), binary.LittleEndian, &config.Ncol)
	}

	buf = readNextChunk(file)
	binary.Read(buf, binary.LittleEndian, &config.UsedIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedFieldIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedCollisionalIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedCollisionalIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedCollisionalIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedCollisionalIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedCollisionalIonize)
	if config.UsedIonize {
		binary.Read(readNextChunk(file), binary.LittleEndian, &config.IonStep)
	}
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.UsedLLDumpingOption)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.UsedLocalSolver)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.RealLx)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.IntSnap)
	binary.Read(readNextChunk(file), binary.LittleEndian, &config.OutputMeshNumber)

	config.TotalOutputMeshNumber = config.OutputMeshNumber[0] * config.OutputMeshNumber[1] * config.OutputMeshNumber[2]
	return config, nil
}

// 設定を表示する
func showConfig(config simulationConfig) {
	fmt.Printf("%+v\n", config)
}
func showPlotConfig(config art) {
	fmt.Println("")
	fmt.Printf("出力先のディレクトリ(テキストファイル) : %s\n", config.OutputASCIIDirectory)
	fmt.Printf("出力先のディレクトリ(VTKファイル))     : %s\n", config.OutputVTKDirectory)
	fmt.Println("出力するデータ")
	for _, v := range config.Field {
		if v.Plot {
			fmt.Printf("%s : %s\n", v.Name, strings.Replace(v.Center, " ", ", ", -1))
		}
	}
}

func loadPlotConfig(v *art) {
	if _, err := os.Stat(plotConfigFileName); os.IsNotExist(err) {
		fmt.Printf("\x1b[35mwarning : %sが存在しないため、新規作成しました。プロットしたいデータを変える場合、%sを変更してください。\n", plotConfigFileName, plotConfigFileName)
		fmt.Printf("Name:データの名前\n")
		fmt.Printf("Plot:出力するか否か\n")
		fmt.Printf("Center:どのデータをプロットするか(複数ある場合はスペース区切りで指定)\x1b[0m\n")

		file, _ := os.Create(plotConfigFileName)
		var buf2 bytes.Buffer
		buf, _ := json.Marshal(*newArt())
		json.Indent(&buf2, buf, "", "  ")
		file.Write(buf2.Bytes())
		file.Close()
	}
	buf, err := ioutil.ReadFile(plotConfigFileName)

	err = json.Unmarshal(buf, v)
	if err != nil {
		fmt.Println(err)
	}
	return
}
func newArt() *art {
	var tempart art
	tempart.Field = append(tempart.Field, subart{"Ex", true, "xy x y"})
	tempart.Field = append(tempart.Field, subart{"Ey", true, "xy x y"})
	tempart.Field = append(tempart.Field, subart{"Ez", false, "xy x y"})
	tempart.Field = append(tempart.Field, subart{"Bx", false, "xy x y"})
	tempart.Field = append(tempart.Field, subart{"By", false, "xy x y"})
	tempart.Field = append(tempart.Field, subart{"Bz", true, "xy x y"})
	tempart.Field = append(tempart.Field, subart{"Jx", false, "xy x y"})
	tempart.Field = append(tempart.Field, subart{"Jy", false, "xy x y"})
	tempart.Field = append(tempart.Field, subart{"Jz", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, subart{"Ion_Density", true, "xy x y"})
	tempart.Particle = append(tempart.Particle, subart{"Ion_Energy", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, subart{"Ion_EnergyFlux_x", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, subart{"Ion_EnergyFlux_y", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, subart{"Electron_Density", true, "xy x y"})
	tempart.Particle = append(tempart.Particle, subart{"Electron_Energy", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, subart{"Electron_EnergyFlux_x", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, subart{"Electron_EnergyFlux_y", false, "xy x y"})
	tempart.OutputASCIIDirectory = "biny_dataASCII"
	tempart.OutputVTKDirectory = "biny_dataVTK"
	return &tempart
}

func main() {
	// 時間を計測用
	start := time.Now()

	// CPU数を取得 macにおけるsysctl -n hw.ncpuの実行結果
	fmt.Println("CPU数", runtime.NumCPU())
	plotConfig = *newArt()
	loadPlotConfig(&plotConfig)
	showPlotConfig(plotConfig)
	// outputDirectoryがあるか確認。なければディレクトリを作成する
	if _, err := os.Stat(plotConfig.OutputASCIIDirectory); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(plotConfig.OutputASCIIDirectory, outputDirectoryAuthority); err != nil {
				fmt.Printf("Error : %sが作れませんでした\n", plotConfig.OutputASCIIDirectory)
				fmt.Println(err)
				os.Exit(-1)
			}
		}
	}
	// outputDirectoryがあるか確認。なければディレクトリを作成する
	if _, err := os.Stat(plotConfig.OutputVTKDirectory); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(plotConfig.OutputVTKDirectory, outputDirectoryAuthority); err != nil {
				fmt.Printf("Error : %sが作れませんでした\n", plotConfig.OutputVTKDirectory)
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
	fmt.Println("")
	fmt.Println("シミュレーションの設定")
	showConfig(config)
	calculateNormalizeConstant(config)

	// snapを終端に達するまで読み込む。
	for fileID := 0; ; fileID++ {
		if loadSnap(file, config, fileID) {
			break
		}
	}

	// 終了時間を記憶
	end := time.Now()
	fmt.Println("正常終了")

	// 終了時間と開始時間の差をとる=かかった時間
	fmt.Println("総時間", end.Sub(start))
}
