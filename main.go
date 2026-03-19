package main

import (
	"os/exec"

	demo "github.com/saschagrunert/demo"
)

func main() {
	d := demo.New()
	d.Add(natsPolicyRun(), "nats", "Use Runtime Enforcer to enforce a security policy on a NATS deployment")
	d.Run()
}

func natsPolicyRun() *demo.Run {
	r := demo.NewRun(
		"Workload Policy Enforcement for NATS",
	)

	r.Setup(cleanupNats)
	r.Cleanup(cleanupNats)

	natsPolicy(r)

	return r
}

func natsPolicy(r *demo.Run) {
	r.Step(demo.S(
		"Install the NATS helm chart and wait for all containers to be ready",
	), demo.S("helm install my-nats nats/nats --wait"))

	r.Step(demo.S(
		"Wait for the WorkloadPolicyProposal to be created",
	), demo.S(
		"kubectl wait --for=create",
		"workloadpolicyproposals.security.rancher.io/sts-my-nats",
		"--timeout=5m",
	))

	r.Step(demo.S(
		"Inspect the spec of the generated WorkloadPolicyProposal",
	), demo.S("kubectl get workloadpolicyproposals.security.rancher.io sts-my-nats -o yaml", "|", "yq '.spec'", "|", "bat -l yaml"))

	r.Step(demo.S(
		"Promote the proposal into an active policy",
	), demo.S(
		"kubectl label workloadpolicyproposals.security.rancher.io -n default sts-my-nats",
		"security.rancher.io/policy-ready=true",
	))

	r.Step(demo.S(
		"Verify the WorkloadPolicy has been created",
	), demo.S("kubectl get workloadpolicies.security.rancher.io sts-my-nats"))

	r.Step(demo.S(
		"Associate the policy with the NATS StatefulSet",
	), demo.S(
		"kubectl patch statefulset my-nats --type='merge'",
		`-p '{"spec":{"template":{"metadata":{"labels":{"security.rancher.io/policy":"sts-my-nats"}}}}}'`,
	))

	r.Step(demo.S(
		"Watch the NATS pod being rolled out with the policy applied",
	), demo.S("kubectl rollout status statefulset/my-nats"))

	r.Step(demo.S(
		"Switch the WorkloadPolicy to protect mode",
	), demo.S(
		"kubectl patch workloadpolicy sts-my-nats --type='merge'",
		`-p '{"spec":{"mode":"protect"}}'`,
	))

	r.Step(demo.S(
		"Verify the WorkloadPolicy is now in protect mode",
	), demo.S("kubectl get workloadpolicies.security.rancher.io sts-my-nats"))

	r.StepCanFail(demo.S(
		"Attempt to run a command not allowed by the policy",
	), demo.S("kubectl exec -ti my-nats-0 -c nats -- date"))

	r.Step(demo.S(
		"Wait for the violation to be scraped and recorded",
	), demo.S(
		"kubectl wait workloadpolicies.security.rancher.io sts-my-nats",
		"--for=jsonpath='{.status.violations}'",
		"--timeout=5m",
	))

	r.Step(demo.S(
		"Show the policy violation recorded in the WorkloadPolicy status",
	), demo.S("kubectl get workloadpolicies.security.rancher.io sts-my-nats -o yaml", "|", "bat -l yaml"))

	r.Step(demo.S(
		"Allow busybox inside the NATS container",
	), demo.S(
		"kubectl patch workloadpolicy sts-my-nats --type='json'",
		`-p='[{"op": "add", "path": "/spec/rulesByContainer/nats/executables/allowed/-", "value": "/bin/busybox"}]'`,
	))

	r.Step(demo.S(
		"Guess what has NOT been removed from this image?!",
	), demo.S("kubectl exec -ti my-nats-0 -c nats -- which apk"))

	r.Step(demo.S(
		"Install curl, like an attacker would do to download their crypto miner",
	), demo.S("kubectl exec -ti my-nats-0 -c nats -- apk add curl"))
}

func cleanupNats() error {
	exec.Command("helm", "uninstall", "my-nats").Run()
	exec.Command("kubectl", "wait", "--for=delete", "pod/my-nats-0", "--timeout=5m").Run()
	exec.Command("kubectl", "delete", "workloadpolicies.security.rancher.io", "sts-my-nats", "--ignore-not-found").Run()
	exec.Command("kubectl", "delete", "workloadpolicyproposals.security.rancher.io", "sts-my-nats", "--ignore-not-found").Run()
	return nil
}
