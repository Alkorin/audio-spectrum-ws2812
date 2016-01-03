package main

//import "fmt"
//import "math"
//import "time"

import "log"
import "os/exec"
import "io"
import "encoding/binary"
import "bytes"
import "github.com/mjibson/go-dsp/fft"
import "github.com/mjibson/go-dsp/window"
import "math/cmplx"
import "net"

const samples = 1024
const finalValues = 128

func main() {
	cmd := exec.Command("parec", "--monitor-stream", "0", "--format=s16le", "--channels=1", "--latency-msec=10")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	vRaw := make([]byte, samples*2)
	vInt := make([]int16, samples)
	vFloat := make([]float64, samples)
	fMag := make([]float64, samples/2)
	data := make([]uint8, finalValues*3)
	for {
		// Read data
		_, err := io.ReadFull(stdout, vRaw)
		// start := time.Now()
		if err != nil {
			break
		}

		// Parse raw values as samples s16le values
		err = binary.Read(bytes.NewReader(vRaw), binary.LittleEndian, &vInt)
		if err != nil {
			log.Fatal(err)
		}

		// Convert them as float values
		for k, v := range vInt {
			vFloat[k] = float64(v)
		}

		// Apply a Hanning window
		window.Apply(vFloat, window.Hann)

		// Do FFT on those values
		f := fft.FFTReal(vFloat)

		// Convert it as magnitude
		for i := 0; i < samples/2; i++ {
			fMag[i] = cmplx.Abs(f[i])
		}

		// Average to 128 values
		for i := 0; i < finalValues; i++ {
			v := float64(0)
			for j := 0; j < 2; j++ {
				v = v + fMag[2*i+j]
			}
			v = v / 1024
			if v > 255 {
				data[3*i] = 255
			} else {
				data[3*i] = uint8(v)
			}
		}

		// Send data over UDP
		saddr, _ := net.ResolveUDPAddr("udp", ":0")
		pudp, _ := net.ListenUDP("udp", saddr)
		pudp.WriteToUDP([]byte(data), &net.UDPAddr{IP: net.IP{10, 0, 0, 10}, Port: 1234})

		// fmt.Printf("Elasped time: %s\n", time.Since(start))
		// fmt.Printf("Sent:\n%+v\n\n", data)
	}
}
