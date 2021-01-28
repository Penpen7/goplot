package energydistribution

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/Penpen7/goplot/cmd/fortbin"
	"github.com/Penpen7/goplot/cmd/physconst"
	"github.com/Penpen7/goplot/cmd/plotconfig"
	"github.com/Penpen7/goplot/cmd/simulationconfig"
)

func writeEnergyDistribution(dltEnergy float32, population []float32, fileName string, wg *sync.WaitGroup) {
	fout, err := os.Create(fileName)
	defer fout.Close()
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(fout)
	writer.WriteString("# energy(eV) population\n")
	for i, v := range population {
		energy := (float32(i+1) - 0.5) * dltEnergy * physconst.NormalizedEnergy
		writer.WriteString(fmt.Sprintln(energy, v))
	}
	writer.Flush()
	wg.Done()
}
func writeLogLogEnergyDistribution(dltEnergy float32, population []float32, fileName string, wg *sync.WaitGroup) {
	fout, err := os.Create(fileName)
	defer fout.Close()
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(fout)
	deltaNumber := float32(10) / float32(len(population))
	writer.WriteString("# energy(eV) population\n")
	for i, v := range population {
		TlogE := float64(i)*float64(deltaNumber) - float64(10)
		energy := float64(float64(i+1)-0.5) * float64(dltEnergy*physconst.NormalizedEnergy) * math.Pow(10.0, TlogE)

		writer.WriteString(fmt.Sprintln(energy, v))
	}
	writer.Flush()
	wg.Done()
}
func LoadWriteEnergyDistribution(file *os.File, config simulationconfig.SimulationConfig, plotConfig plotconfig.Art, fileID int, wg *sync.WaitGroup) {
	for i := int32(1); i <= config.TotalParticleSpecies; i++ {
		var averageChargeRate, averageEnergy, dltEnergy, Eimaxt float32
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &averageChargeRate)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &averageEnergy)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &dltEnergy)
		population := make([]float32, config.MomentumMeshNumber)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &population)
		if i <= config.IonNumber {
			if isfound := plotconfig.SearchSubart(plotConfig.Particle, "Ion_Energy_Distribution"); isfound {
				wg.Add(1)
				go writeEnergyDistribution(dltEnergy, population, fmt.Sprintf("%s/Ion_Energy_Distribution%04d_is=%02d.txt", plotConfig.OutputASCIIDirectory, fileID, i), wg)
			}
		} else {
			if isfound := plotconfig.SearchSubart(plotConfig.Particle, "Electron_Energy_Distribution"); isfound {
				wg.Add(1)
				go writeEnergyDistribution(dltEnergy, population, fmt.Sprintf("%s/Electron_Energy_Distribution%04d_is=%02d.txt", plotConfig.OutputASCIIDirectory, fileID, i), wg)
			}
		}
		fortbin.ReadNextChunk(file) //FF2
		fortbin.ReadNextChunk(file) //FF3

		// log-log
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &Eimaxt)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &population)
		fortbin.ReadNextChunk(file) //FF2
		fortbin.ReadNextChunk(file) //FF3
		if i <= config.IonNumber {
			if isfound := plotconfig.SearchSubart(plotConfig.Particle, "Ion_Energy_DistributionLogLog"); isfound {
				wg.Add(1)
				go writeEnergyDistribution(dltEnergy, population, fmt.Sprintf("%s/Ion_Energy_DistributionLog%04d_is=%02d.txt", plotConfig.OutputASCIIDirectory, fileID, i), wg)
			}
		} else {
			if isfound := plotconfig.SearchSubart(plotConfig.Particle, "Electron_Energy_DistributionLogLog"); isfound {
				wg.Add(1)
				go writeEnergyDistribution(dltEnergy, population, fmt.Sprintf("%s/Electron_Energy_DistributionLog%04d_is=%02d.txt", plotConfig.OutputASCIIDirectory, fileID, i), wg)
			}
		}
	}
	for i := int32(1); i <= config.TotalParticleSpecies; i++ {
		fortbin.ReadNextChunk(file)
		fortbin.ReadNextChunk(file)
		fortbin.ReadNextChunk(file)
		fortbin.ReadNextChunk(file)
		fortbin.ReadNextChunk(file)
		fortbin.ReadNextChunk(file)
		fortbin.ReadNextChunk(file)
		fortbin.ReadNextChunk(file)
		fortbin.ReadNextChunk(file)
		fortbin.ReadNextChunk(file)
		fortbin.ReadNextChunk(file)
		fortbin.ReadNextChunk(file)
	}
}
