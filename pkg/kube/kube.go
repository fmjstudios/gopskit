package kube

import "fmt"

// ref: https://github.com/iximiuz/client-go-examples/blob/main/kubeconfig-from-yaml/main.go
// ref: https://github.com/a4abhishek/Client-Go-Examples/blob/master/exec_to_pod/exec_to_pod.go
// ref: https://miminar.fedorapeople.org/_preview/openshift-enterprise/registry-redeploy/go_client/executing_remote_processes.html

func Log(message string) {
	fmt.Println(message)
}
