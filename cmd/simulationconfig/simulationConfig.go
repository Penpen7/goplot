package simulationconfig

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/Penpen7/goplot/cmd/fortbin"
)

type SimulationParticleConfig struct {
	N_p                            int32
	Np                             int32
	Nps                            int32
	ParticleMass                   float64
	ParticleCharge                 float64
	ParticleTempretureFunction     int32
	ParticleTempreture             float64
	ParticleOutGoing               [3]bool
	DensityFunctionType            string
	Rns_b                          float64
	LoadType                       int32
	NxFunc                         int32
	NyFunc                         int32
	Nix                            [4]float32
	Niy                            [4]float32
	Rds                            float64
	ClusterLoadingOption           int32
	ClusterShape                   int32
	NumberCluster                  int32
	Xclr                           [2]float64
	Yclr                           [2]float64
	ClusterDistance                float64
	LLDumping                      bool
	Atom                           string
	ParticleInitialChargeForIonize float64
}
type SimulationLaserConfig struct {
	RLw          float64
	X0           float64
	X1           float64
	Y1           float64
	RLx          float64
	RLy          float64
	E0           float64
	A0_0         float64
	Tau0         float64
	T_0          float64
	Lambda       float64
	Dy0          float64
	LaserFocus   bool
	FocusLength  float64
	ExternalCrnt float64
	EStc         [3]float64
	IsLaserRise  bool
	Polarize     string
	Direction    int32
}
type SimulationConfig struct {
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
	Particle                   []SimulationParticleConfig
	Laser                      SimulationLaserConfig
}

func LoadSetting(fname string) (SimulationConfig, error) {
	var config SimulationConfig
	file, err := os.Open(fname)
	if err != nil {
		return config, err
	}
	buf := fortbin.ReadNextChunk(file)
	config.Version = strings.TrimSpace(fmt.Sprintf("%s", buf.Bytes()))
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.ParallelNumber)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Dimension)

	buf = fortbin.ReadNextChunk(file)
	binary.Read(buf, binary.LittleEndian, &config.VelocityLight)
	binary.Read(buf, binary.LittleEndian, &config.DeltTime)
	binary.Read(buf, binary.LittleEndian, &config.DeltX)

	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.SystemL)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.AverageDensity)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.MeshNumber)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.FildBoundaryCondition)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.TotalParticleNumber)

	buf = fortbin.ReadNextChunk(file)
	binary.Read(buf, binary.LittleEndian, &config.TotalParticleSpecies)
	binary.Read(buf, binary.LittleEndian, &config.IonNumber)
	binary.Read(buf, binary.LittleEndian, &config.ElectronNumber)

	config.Loadtype = make([]int32, config.TotalParticleSpecies)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Loadtype)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.ClusterOption)
	if config.ClusterOption {
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.ClusterNumber)
	}
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.CollisionOption)
	if config.CollisionOption {
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Ncol)
	}

	buf = fortbin.ReadNextChunk(file)
	binary.Read(buf, binary.LittleEndian, &config.UsedIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedFieldIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedCollisionalIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedCollisionalIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedCollisionalIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedCollisionalIonize)
	binary.Read(buf, binary.LittleEndian, &config.UsedCollisionalIonize)
	if config.UsedIonize {
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.IonStep)
	}
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.UsedLLDumpingOption)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.UsedLocalSolver)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.RealLx)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.IntSnap)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.OutputMeshNumber)
	buf = fortbin.ReadNextChunk(file)
	binary.Read(buf, binary.LittleEndian, &config.MomentumMeshNumber)
	binary.Read(buf, binary.LittleEndian, &config.SpaceMeshNumberForMomentum)
	config.Particle = make([]SimulationParticleConfig, config.TotalParticleSpecies)
	for ionID := int32(0); ionID < config.IonNumber; ionID++ {
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].LoadType)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].N_p)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].Np)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].Nps)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].ParticleMass)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].ParticleCharge)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].ParticleTempretureFunction)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].ParticleTempreture)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].Rns_b)

		if config.Particle[ionID].LoadType == 0 {
			bufstr := make([]byte, 4)
			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &bufstr)
			config.Particle[ionID].DensityFunctionType = strings.TrimSpace(fmt.Sprintf("%s", bufstr))
			if config.Particle[ionID].DensityFunctionType == "x" {
				binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].NxFunc)
				binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].Nix)
			} else if config.Particle[ionID].DensityFunctionType == "y" {
				binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].NyFunc)
				binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].Niy)
			}
		} else if config.Particle[ionID].LoadType == 1 {
			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].Rds)
			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].ClusterLoadingOption)
			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].ClusterShape)
			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].NumberCluster)
			buf := fortbin.ReadNextChunk(file)
			binary.Read(buf, binary.LittleEndian, &config.Particle[ionID].Xclr)
			binary.Read(buf, binary.LittleEndian, &config.Particle[ionID].Yclr)

			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].ClusterDistance)
		}
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].LLDumping)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].ParticleOutGoing[0])
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].ParticleOutGoing[1])
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].ParticleOutGoing[2])
		if config.UsedIonize {
			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].Atom)
			binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[ionID].ParticleInitialChargeForIonize)

		}
	}
	for electronID := config.IonNumber; electronID < config.TotalParticleSpecies; electronID++ {
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].N_p)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].Np)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].Nps)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].ParticleMass)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].ParticleCharge)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].ParticleTempretureFunction)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].ParticleTempreture)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].Rns_b)

		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].LLDumping)
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].ParticleOutGoing[0])
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].ParticleOutGoing[1])
		binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Particle[electronID].ParticleOutGoing[2])
	}
	fortbin.ReadNextChunk(file)
	buf = fortbin.ReadNextChunk(file)
	binary.Read(buf, binary.LittleEndian, &config.Laser.RLw)
	binary.Read(buf, binary.LittleEndian, &config.Laser.X0)
	binary.Read(buf, binary.LittleEndian, &config.Laser.X1)
	binary.Read(buf, binary.LittleEndian, &config.Laser.Y1)
	binary.Read(buf, binary.LittleEndian, &config.Laser.RLx)
	binary.Read(buf, binary.LittleEndian, &config.Laser.RLy)
	binary.Read(buf, binary.LittleEndian, &config.Laser.E0)

	buf = fortbin.ReadNextChunk(file)
	var bufBool []byte
	bufBool = make([]byte, 4)
	binary.Read(buf, binary.LittleEndian, &bufBool)
	binary.Read(bytes.NewReader(bufBool), binary.LittleEndian, &config.Laser.IsLaserRise)
	var bufstr []byte = make([]byte, 4)
	binary.Read(buf, binary.LittleEndian, &bufstr)
	config.Laser.Polarize = strings.TrimSpace(fmt.Sprintf("%s", bufstr))
	binary.Read(buf, binary.LittleEndian, &config.Laser.Direction)

	buf = fortbin.ReadNextChunk(file)
	binary.Read(buf, binary.LittleEndian, &config.Laser.A0_0)
	binary.Read(buf, binary.LittleEndian, &config.Laser.Tau0)
	binary.Read(buf, binary.LittleEndian, &config.Laser.T_0)
	binary.Read(buf, binary.LittleEndian, &config.Laser.Lambda)
	binary.Read(buf, binary.LittleEndian, &config.Laser.Dy0)

	buf = fortbin.ReadNextChunk(file)
	binary.Read(buf, binary.LittleEndian, &config.Laser.LaserFocus)
	binary.Read(buf, binary.LittleEndian, &config.Laser.FocusLength)
	binary.Read(fortbin.ReadNextChunk(file), binary.LittleEndian, &config.Laser.ExternalCrnt)

	buf = fortbin.ReadNextChunk(file)
	binary.Read(buf, binary.LittleEndian, &config.Laser.EStc)
	config.TotalOutputMeshNumber = config.OutputMeshNumber[0] * config.OutputMeshNumber[1] * config.OutputMeshNumber[2]
	return config, nil
}

// 設定を表示する
func ShowConfig(config SimulationConfig) {
	fmt.Printf("%+v\n", config)
}
