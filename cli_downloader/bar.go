package main

import (
	"time"

	"github.com/cheggaaa/pb/v3"
)

func ProgreeBar() {

	count := 100000

	// create and start new bar
	bar := pb.StartNew(count)

	// start bar from 'default' template
	// bar := pb.Default.Start(count)

	// start bar from 'simple' template
	// bar := pb.Simple.Start(count)

	// start bar from 'full' template
	// bar := pb.Full.Start(count)

	for i := 0; i < count; i++ {
		bar.Increment()
		time.Sleep(time.Millisecond)
	}

	// finish bar
	bar.Finish()
}
