package physconst

import (
	"math"

	"github.com/Penpen7/goplot/cmd/simulationconfig"
)

var ElectricFieldNormalizeConstant float32
var MagneticFieldNormalizeConstant float32
var NormalizedEnergy float32

func CalculateNormalizeConstant(sc simulationconfig.SimulationConfig) {
	lightSpeed := 2.99792458e+10     //c_r
	electronMass := 9.10938356e-28   //rme_r
	electricUnit := 4.8032e-10       //e_r
	electronVoltToJoule := 1.602e-12 //eV_J_r
	normalizedDeltaX := sc.RealLx / sc.SystemL[0]
	normalizedPlasmaFrequency := lightSpeed / sc.VelocityLight / normalizedDeltaX
	normalizedNumberDensity := math.Pow(normalizedPlasmaFrequency, 2) * electronMass / (4.0 * math.Pi * math.Pow(electricUnit, 2))
	ElectricFieldNormalizeConstant = float32(4.0 * math.Pi * normalizedNumberDensity * electricUnit * normalizedDeltaX * 1e+4 * 3.0)
	MagneticFieldNormalizeConstant = ElectricFieldNormalizeConstant / float32(lightSpeed*1e-2)
	NormalizedEnergy = float32(4.0 * math.Pi * normalizedNumberDensity * electricUnit * electricUnit * normalizedDeltaX * normalizedDeltaX / electronVoltToJoule)
}
