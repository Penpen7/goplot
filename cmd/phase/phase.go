package phase

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"sync"

	"github.com/Penpen7/goplot/cmd/fortbin"
	"github.com/Penpen7/goplot/cmd/plotconfig"
	"github.com/Penpen7/goplot/cmd/simulationconfig"
	"github.com/Penpen7/goplot/cmd/utility"
)

func writePhaseSpace(xdata []float32, ydata []float32, vdata [][]float32, fname string, wg *sync.WaitGroup) {
	fout, err := os.Create(fname)
	defer fout.Close()
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(fout)
	for xindex, x := range xdata {
		for yindex, y := range ydata {
			writer.WriteString(fmt.Sprintln(x, y, vdata[xindex][yindex]))
		}
		writer.WriteString(fmt.Sprintf("\n"))
	}
	writer.Flush()
	wg.Done()
}

func LoadWritePhaseSpace(file *os.File, config simulationconfig.SimulationConfig, plotConfig plotconfig.Art, fileID int, wg *sync.WaitGroup) {
	momentum_title := [...]string{"pxpy", "pypz", "pzpx"}
	position_title := [...]string{"xpx", "xpy", "xpz", "ypx", "ypy", "ypz"}
	velocity_title := [...]string{"vxvy", "vyvz", "vzvx"}
	position_velocity_title := [...]string{"xvx", "xvy", "xvz", "yvx", "yvy", "yvz"}

	for iparticle := int32(1); iparticle <= config.TotalParticleSpecies; iparticle++ {
		var dltmomentum float32
		momentumvsmomentum := []float32{}
		momentumvsmomentum = make([]float32, config.MomentumMeshNumber*config.MomentumMeshNumber)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &dltmomentum)
		momentum := make([]float32, config.MomentumMeshNumber)
		for i, _ := range momentum {
			momentum[i] = float32(dltmomentum) * (float32(int32(i)-config.MomentumMeshNumber/2) - 0.5) / float32(config.Particle[iparticle-1].ParticleMass*config.VelocityLight)
		}

		for _, v := range momentum_title {
			fmt.Printf("\r\033[K loading... %s", v)
			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &momentumvsmomentum)

			wg.Add(1)
			go writePhaseSpace(momentum, momentum, utility.Slice1Dto2D(momentumvsmomentum, config.MomentumMeshNumber, config.MomentumMeshNumber),
				fmt.Sprintf("%s/%s%04d_is=%02d.txt", plotConfig.OutputASCIIDirectory, v, fileID, iparticle), wg)
		}

		for titlei, v := range position_title {
			fmt.Printf("\r\033[K loading... %s", v)
			positionvsmomentum := make([]float32, config.OutputMeshNumber[titlei/3]*config.MomentumMeshNumber)
			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &positionvsmomentum)
			position := make([]float32, config.OutputMeshNumber[titlei/3])
			for iposition := int32(0); iposition < config.OutputMeshNumber[titlei/3]; iposition++ {
				position[iposition] = float32(iposition)
			}
			wg.Add(1)
			var buf [][]float32
			if titlei/3 == 1 {
				buf = utility.Transpy(positionvsmomentum, int(config.OutputMeshNumber[1]), int(config.ParallelNumber), int(config.MomentumMeshNumber))
			} else {
				buf = utility.Slice1Dto2D(positionvsmomentum, config.OutputMeshNumber[titlei/3], config.MomentumMeshNumber)
			}
			go writePhaseSpace(position, momentum, buf,
				fmt.Sprintf("%s/%s%04d_is=%02d.txt", plotConfig.OutputASCIIDirectory, v, fileID, iparticle), wg)
		}

		fortbin.ReadNextChunk(file)
		for i := int32(0); i < int32(len(velocity_title)); i++ {
			fortbin.ReadNextChunk(file)
		}
		for i := int32(0); i < int32(len(position_velocity_title)); i++ {
			fortbin.ReadNextChunk(file)
		}
	}
}
