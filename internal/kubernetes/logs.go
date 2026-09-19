package kubernetes

import (
	"bytes"
	"context"
	"io"

	corev1 "k8s.io/api/core/v1"
)

const DefaultLogTail int64 = 80

func (c *Client) PodLogs(ctx context.Context, namespace, pod, container string, tail int64) (string, error) {
	if tail <= 0 {
		tail = DefaultLogTail
	}
	opts := &corev1.PodLogOptions{
		Container: container,
		TailLines: &tail,
	}
	req := c.Kube.CoreV1().Pods(namespace).GetLogs(pod, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, stream); err != nil {
		return "", err
	}
	return buf.String(), nil
}
