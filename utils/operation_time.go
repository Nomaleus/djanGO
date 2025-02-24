package utils

import (
	"math/rand/v2"
	"runtime"
	"time"
)

func GetOperationTime(operation string) int {
	baseTime := map[string]int{
		"+": 1000,
		"-": 1000,
		"*": 2000,
		"/": 3000,
	}

	numCPU := runtime.NumCPU()

	var cpuUsage float64
	start := time.Now()
	runtime.GC()
	elapsed := time.Since(start)

	cpuUsage = float64(elapsed.Milliseconds()) / 1000.0
	if cpuUsage > 1.0 {
		cpuUsage = 1.0
	}

	adjustedTime := float64(baseTime[operation]) * (1 + cpuUsage) / float64(numCPU)

	randomFactor := 0.9 + rand.Float64()*0.2
	finalTime := int(adjustedTime * randomFactor)

	if finalTime < 100 {
		finalTime = 100
	}

	return finalTime
}
