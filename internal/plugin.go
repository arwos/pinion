/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a GPL-3.0 license that can be found in the LICENSE file.
 */

package internal

import (
	"go.arwos.org/pinion/internal/config"
	"go.arwos.org/pinion/internal/httpsrv"
	"go.osspkg.com/goppy/v2/plugins"
)

var Plugin = plugins.Plugins{
	plugins.Plugin{
		Config: &config.Config{},
		Inject: httpsrv.New,
	},
}
