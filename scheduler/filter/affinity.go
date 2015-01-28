package filter

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/samalba/dockerclient"
)

// AffinityFilter selects only nodes based on other containers on the node.
type AffinityFilter struct {
}

func (f *AffinityFilter) Filter(config *dockerclient.ContainerConfig, nodes []*cluster.Node) ([]*cluster.Node, error) {
	affinities, err := parseExprs("affinity", config.Env)
	if err != nil {
		return nil, err
	}

	for _, affinity := range affinities {
		log.Debugf("matching affinity: %s%s%s", affinity.key, OPERATORS[affinity.operator], affinity.value)

		candidates := []*cluster.Node{}
		for _, node := range nodes {
			switch affinity.key {
			case "container":
				var whats []string
				for _, container := range node.Containers() {
					whats = append(whats, container.Id, strings.TrimPrefix(container.Names[0], "/"))
				}
				if affinity.Match(whats...) {
					candidates = append(candidates, node)
				}
			case "image":
				var whats []string
				for _, image := range node.Images() {
					whats = append(whats, image.Id)
					for _, tag := range image.RepoTags {
						whats = append(whats, tag, strings.Split(tag, ":")[0])
					}
				}
				if affinity.Match(whats...) {
					candidates = append(candidates, node)
				}
			}
		}
		if len(candidates) == 0 {
			return nil, fmt.Errorf("unable to find a node that satisfies %s%s%s", affinity.key, OPERATORS[affinity.operator], affinity.value)
		}
		nodes = candidates
	}
	return nodes, nil
}
