package modelers

import "wutka.com/icfpc/builder"

type Modeler interface {
	Model(modelBytes []byte, b *builder.Bot)
}
