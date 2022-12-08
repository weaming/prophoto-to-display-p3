package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"runtime"
	"sync"

	C "github.com/weaming/go-chromath"
	"golang.org/x/image/tiff"
)

func main() {
	input := ""
	output := ""
	switch len(os.Args) {
	case 1:
		fmt.Println("missing file path!")
		fmt.Println("Usage: prophoto-rgb-to-apple-display-p3 INPUT OUTPUT")
		os.Exit(0)
	case 2:
		input = os.Args[1]
		output = os.Args[1] + ".jpg"
	default:
		input = os.Args[1]
		output = os.Args[2]
	}

	fileIn, err := os.Open(input)
	if err != nil {
		panic(err)
	}

	config, format, err := image.DecodeConfig(fileIn)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v: %vx%v\n", format, config.Width, config.Height)

	img, err := tiff.Decode(fileIn)
	if err != nil {
		panic(err)
	}

	fileOut, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	defer fileOut.Close()

	wg := sync.WaitGroup{}
	limit := make(chan int, runtime.NumCPU()*1000)
	bounds := img.Bounds()
	imgOut := image.NewRGBA(image.Rect(0, 0, bounds.Max.X, bounds.Max.Y))
	total := config.Width * config.Height
	percent := 0.0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			limit <- 1
			wg.Add(1)
			if float64(x*y)/float64(total) > percent {
				fmt.Printf("\r%v %v%%", output, int(percent*100))
				percent += 0.05
			}

			go func(x, y int) {
				defer func() {
					wg.Done()
					<-limit
				}()

				color := img.At(x, y)
				offset := (y * imgOut.Stride) + x*4
				r, g, b, a := color.RGBA()
				color2 := Convert(C.RGB{float64(r), float64(g), float64(b)})
				color3 := [3]uint8{uint8(color2[0] * 255 / 65535), uint8(color2[1] * 255 / 65535), uint8(color2[2] * 255 / 65535)}
				// fmt.Println(color, color2, color3, offset/4)
				imgOut.Pix[offset+0] = color3[0]
				imgOut.Pix[offset+1] = color3[1]
				imgOut.Pix[offset+2] = color3[2]
				imgOut.Pix[offset+3] = uint8(a)
			}(x, y)
		}
	}
	fmt.Println("")

	wg.Wait()
	quality := jpeg.Options{100}
	jpeg.Encode(fileOut, imgOut, &quality)
	fmt.Println(ExecGetOutput(fmt.Sprintf(`exiftool "-icc_profile<=DisplayP3.icc" -overwrite_original %v`, output)))
}

func Convert(point C.RGB) C.RGB {
	// 16 bits ProtPhoto RGB
	pp2xyz := C.NewRGBTransformer(&C.SpaceProPhotoRGB, &C.AdaptationBradford, nil, &C.Scaler16bClamping, 1.0, nil)
	// Display P3 RGB
	dp2xyz := C.NewRGBTransformer(&C.SpaceDisplayP3RGB, &C.AdaptationBradford, nil, &C.Scaler16bClamping, 1.0, nil)
	return dp2xyz.Invert(pp2xyz.Convert(point))
}

func ExecGetOutput(command string) string {
	cmd := exec.Command("bash", []string{"-c", command}...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(`execute command "%v" error: %v\n`, command, err)
	}
	return string(stdoutStderr)
}
