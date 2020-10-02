/*MIT License
-----------

Copyright (c) 2020 Steve McDaniel, Corey Gaspard

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation
files (the "Software"), to deal in the Software without
restriction, including without limitation the rights to use,
copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following
conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/

package controller

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	pb "gitlab.com/skyhuborg/proto-tracker-controller-go"
)

func (s *Server) getContainerLog(containerName string) string {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	options := types.ContainerLogsOptions{ShowStdout: true}
	// Replace this ID with a container that really exists
	out, err := cli.ContainerLogs(ctx, containerName, options)
	if err != nil {
		panic(err)
	}

	buf := new(strings.Builder)
	io.Copy(buf, out)
	return buf.String()
}

func (s *Server) listContainers() []pb.Container {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	containersAry := []pb.Container{}

	// message Container {
	// 	string id = 1;
	// 	string name = 2;
	// 	string image = 3;
	// 	string status = 4;
	// }

	for _, container := range containers {
		newContainer := pb.Container{}
		newContainer.Id = container.ID[:10]
		newContainer.Name = container.Names[0]
		newContainer.Image = container.Image
		newContainer.Status = container.Status
		containersAry = append(containersAry, newContainer)
		//fmt.Printf("%s %s %s\n", container.ID[:10], container.Names, container.Image)
	}

	return containersAry
}
