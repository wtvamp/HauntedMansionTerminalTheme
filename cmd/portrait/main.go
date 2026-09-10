// Command portrait converts an image into ASCII art for content/ghosts/.
//
//	portrait leota.png                              # 60 columns, to stdout
//	portrait -w 72 -ramp dense leota.png            # wider, finer shading
//	portrait -w 56 -o content/ghosts/leota.txt p.jpg
//	portrait -gamma 1.4 dark-photo.png              # lift the shadows
//
// Output is plain ASCII with no escape sequences, so it drops straight into
// content/ghosts/ and works in the greeting.
package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"strings"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/portrait"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "portrait:", err)
		os.Exit(1)
	}
}

func run() error {
	o := portrait.Defaults()
	flag.IntVar(&o.Width, "w", o.Width, "output width in terminal columns")
	flag.StringVar(&o.Ramp, "ramp", o.Ramp, "character ramp: "+strings.Join(portrait.RampNames, ", "))
	flag.StringVar(&o.RampChars, "ramp-chars", "", "literal ramp, darkest first; overrides -ramp")
	flag.Float64Var(&o.Aspect, "aspect", o.Aspect, "terminal cell height:width ratio")
	flag.Float64Var(&o.Gamma, "gamma", o.Gamma, "above 1 lifts shadows, below 1 deepens them")
	flag.BoolVar(&o.Normalize, "normalize", o.Normalize, "stretch the image's brightness range to full before mapping")
	flag.Float64Var(&o.Black, "black", o.Black, "input level mapped to the darkest ramp character")
	flag.Float64Var(&o.White, "white", o.White, "input level mapped to the brightest ramp character")
	flag.BoolVar(&o.Invert, "invert", o.Invert, "swap light and dark, for a light background")
	flag.BoolVar(&o.Trim, "trim", o.Trim, "drop blank edge rows and columns")
	out := flag.String("o", "", "write here instead of stdout")
	regionsPath := flag.String("regions", "", "companion region image; enables the colour map")
	mapOut := flag.String("map", "", "write the colour map here (requires -regions)")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		return fmt.Errorf("need exactly one image file")
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		return err
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("decode %s: %w (png, jpeg and gif are supported)", flag.Arg(0), err)
	}

	var art, colorMap string
	if *regionsPath != "" {
		rf, err := os.Open(*regionsPath)
		if err != nil {
			return err
		}
		defer rf.Close()
		rimg, _, err := image.Decode(rf)
		if err != nil {
			return fmt.Errorf("decode regions %s: %w", *regionsPath, err)
		}
		art, colorMap, err = portrait.ConvertWithRegions(img, rimg, o)
		if err != nil {
			return err
		}
	} else {
		if *mapOut != "" {
			return fmt.Errorf("-map needs -regions: the colour map comes from the region image")
		}
		art, err = portrait.Convert(img, o)
		if err != nil {
			return err
		}
	}

	if *mapOut != "" {
		if err := os.WriteFile(*mapOut, []byte(colorMap), 0o644); err != nil {
			return err
		}
	}

	if *out == "" {
		fmt.Print(art)
		return nil
	}
	if err := os.WriteFile(*out, []byte(art), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s: %s -> %s, %d columns\n", format, flag.Arg(0), *out, portrait.Width(art))
	return nil
}
