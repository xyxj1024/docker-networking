package exec

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

func GenerateSnapshot(ctx context.Context, config cache.SnapshotCache) {
	num := len(config.GetStatusKeys())
	logrus.Infof("%d connected nodes\n", num)

	if num > 0 {
		for i := 0; i < num; i++ {
			version := time.Now().Format(time.RFC3339) // timestamp as version number
			nodeId := config.GetStatusKeys()[i]        // Use IP address as node ID
			logrus.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot " + fmt.Sprint(version) +
				" for nodeID " + fmt.Sprint(nodeId))
		}
	}
}
