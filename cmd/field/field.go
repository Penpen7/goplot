package field

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Penpen7/goplot/cmd/fortbin"
	"github.com/Penpen7/goplot/cmd/physconst"
	"github.com/Penpen7/goplot/cmd/plotconfig"
	"github.com/Penpen7/goplot/cmd/simulationconfig"
	"github.com/Penpen7/goplot/cmd/utility"
)

func WriteFieldData(g [][][]float32, mode string, fname string, wg *sync.WaitGroup) {
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
	case "whole_average":
    sum := float32(0)
    for x := 0; x < xsize; x++{
      for y:= 0; y < ysize; y++ {
        for z:=0; z < zsize; z++ {
          sum += g[x][y][z]
        }
      }
    }
    average := float32(sum) / float32(xsize * ysize * zsize)
    writer.WriteString(fmt.Sprintln(average))
    break
	default:
		fmt.Println("Warning:invalid mode:", mode)
	}
	writer.Flush()
	fout.Close()
	wg.Done()
}
func WriteFieldVTK(g [][][]float32, fname string, arrayName string, config simulationconfig.SimulationConfig, wg *sync.WaitGroup) {
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
func LoadWriteFieldData(file *os.File, config simulationconfig.SimulationConfig, plotConfig plotconfig.Art, fileID int, wg *sync.WaitGroup) {
	title := [...]string{"Ex", "Ey", "Ez", "Bx", "By", "Bz", "Jx", "Jy", "Jz"}
	normalizeConst := [...]float32{physconst.ElectricFieldNormalizeConstant, physconst.ElectricFieldNormalizeConstant, physconst.ElectricFieldNormalizeConstant,
		physconst.MagneticFieldNormalizeConstant, physconst.MagneticFieldNormalizeConstant, physconst.MagneticFieldNormalizeConstant,
		1.0, 1.0, 1.0}
	for i, v := range title {
		fmt.Printf("\r\033[K loading... %s", v)
		g := []float32{}
		g = make([]float32, config.TotalOutputMeshNumber)
		nextchunk := fortbin.ReadNextChunk(file)
		binary.Read(nextchunk, binary.LittleEndian, &g)
		buf := utility.Slice1Dto3D(g, config.OutputMeshNumber[0], config.OutputMeshNumber[1], config.OutputMeshNumber[2], normalizeConst[i])

		for _, vconfig := range plotConfig.Field {
			if vconfig.Name == v {
				if !vconfig.Plot {
					break
				}
				for _, vcenter := range strings.Split(vconfig.Center, " ") {
					wg.Add(1)
					if vcenter == "vtk" {
						go WriteFieldVTK(buf, fmt.Sprintf("%s/%s%04d.vti", plotConfig.OutputVTKDirectory, v, fileID), v, config, wg)
					} else {
						go WriteFieldData(buf, vcenter, fmt.Sprintf("%s/%s_%s_%04d.txt", plotConfig.OutputASCIIDirectory, v, vcenter, fileID), wg)
					}
				}
			}
		}
	}
}
func LoadWriteParticleMeshData(file *os.File, config simulationconfig.SimulationConfig, plotConfig plotconfig.Art, fileID int, wg *sync.WaitGroup) {

	title_particle := [...]string{"Ion_Density", "Ion_Energy", "Ion_EnergyFlux_x", "Ion_EnergyFlux_y"}
	title_particle_Electron := [...]string{"Electron_Density", "Electron_Energy", "Electron_EnergyFlux_x", "Electron_EnergyFlux_y"}

	for ionID := int32(1); ionID <= config.IonNumber; ionID++ {
		for _, v := range title_particle {
			fmt.Printf("\r\033[K loading... %s", v)
			g := []float32{}
			g = make([]float32, config.TotalOutputMeshNumber)
			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &g)
			buf := utility.Slice1Dto3D(g, config.OutputMeshNumber[0], config.OutputMeshNumber[1], config.OutputMeshNumber[2], 1.0)

			for _, vconfig := range plotConfig.Particle {
				if vconfig.Name == v && vconfig.Plot {
					for _, vplot := range strings.Split(vconfig.Center, " ") {
						wg.Add(1)
						if vplot == "vtk" {
							go WriteFieldVTK(buf, fmt.Sprintf("%s/%s%04d_is=%02d.vti", plotConfig.OutputVTKDirectory, v, fileID, ionID), v, config, wg)
						} else {
							go WriteFieldData(buf, vplot, fmt.Sprintf("%s/%s_%s_%04d_is=%02d.txt", plotConfig.OutputASCIIDirectory, v, vplot, fileID, ionID), wg)
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
			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &g)
			buf := utility.Slice1Dto3D(g, config.OutputMeshNumber[0], config.OutputMeshNumber[1], config.OutputMeshNumber[2], 1.0)
			for _, vconfig := range plotConfig.Particle {
				if vconfig.Name == v && vconfig.Plot {
					for _, vplot := range strings.Split(vconfig.Center, " ") {
						wg.Add(1)
						if vplot == "vtk" {
							go WriteFieldVTK(buf, fmt.Sprintf("%s/%s%04d_is=%02d.vti", plotConfig.OutputVTKDirectory, v, fileID, ElectronID), v, config, wg)
						} else {
							go WriteFieldData(buf, vplot, fmt.Sprintf("%s/%s_%s_%04d_is=%02d.txt", plotConfig.OutputASCIIDirectory, v, vplot, fileID, ElectronID), wg)
						}
					}
				}
			}
		}
	}
}
