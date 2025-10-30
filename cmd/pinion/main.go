/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a GPL-3.0 license that can be found in the LICENSE file.
 */

package main

import (
	"go.osspkg.com/goppy/v2"

	"go.arwos.org/pinion/internal"
)

func main() {
	app := goppy.New("pinion", "v0.0.0-dev", "http proxy server")
	app.Plugins(internal.Plugin...)
	app.Run()
}
