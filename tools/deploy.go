package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

const (
	schedev       = "schedev"
	kc_sched      = "kc-sched"
	schedrox      = "schedrox"
	InstanceClass = "InstanceClass"
	App           = "App"
	F4            = "F4"
	F1            = "F1"
	optional      = "optional"
	always        = "always"
	Secure        = "Secure"
)

var (
	validInstanceClasses = map[string]bool{
		"F1": true,
		"F2": true,
		"F4": true,
		"":   true,
	}

	defaultInstanceClass = map[string]string{
		schedev:  F1,
		kc_sched: F4,
		schedrox: F1,
	}

	defaultSecure = map[string]string{
		schedev:  optional,
		kc_sched: always,
		schedrox: always,
	}

	defaultCronBackup = map[string]string{
		kc_sched: "kc-sched-datastore-backup",
	}

	kinds = []string{
		"Kind",
		"Change",
		"Attest",
		"TimeReport",
		"Term",
		"Event",
		"Participant",
		"EventWeek",
		"Contact",
		"Auth",
		"RequiredParticipantType",
		"DomainUser",
		"User",
		"Role",
		"EventType",
		"CustomFilter",
		"IsReported",
		"UserPropertyForUser",
		"EventKind",
		"SalaryConfig",
		"ParticipantType",
		"UserPropertyForDomain",
		"Domain",
		"Location",
	}
)

func createAppYaml(app string, instanceClass string) {
	dst, err := os.Create("app.yaml")
	if err != nil {
		panic(err)
	}
	ctx := map[string]interface{}{
		InstanceClass: F1,
	}
	if def, found := defaultInstanceClass[app]; found {
		ctx[InstanceClass] = def
	}
	if instanceClass != "" {
		ctx[InstanceClass] = instanceClass
	}
	if def, found := defaultSecure[app]; found {
		ctx[Secure] = def
	}
	tmpl, err := template.ParseFiles("app.yaml.template")
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(dst, ctx); err != nil {
		panic(err)
	}
	if err := dst.Close(); err != nil {
		panic(err)
	}
}

func updateCronYaml(app string) {
	if bucketName := defaultCronBackup[app]; bucketName != "" {
		kindStrings := []string{}
		for _, kind := range kinds {
			kindStrings = append(kindStrings, fmt.Sprintf("kind=%s", kind))
		}
		if err := ioutil.WriteFile("cron.yaml", []byte(fmt.Sprintf(`cron:
- description: My Daily Backup
  url: /_ah/datastore_admin/backup.create?name=BackupToCloud&%s&kind=EventLog&filesystem=gs&gs_bucket_name=%s
  schedule: every 24 hours
  target: ah-builtin-python-bundle
`, strings.Join(kindStrings, "&"), bucketName)), 0666); err != nil {
			panic(err)
		}
	} else {
		if err := ioutil.WriteFile("cron.yaml", []byte("cron:"), 0666); err != nil {
			panic(err)
		}
	}
}

func deployApp(app string, instanceClass string) {
	createAppYaml(app, instanceClass)
	updateCronYaml(app)
	cmd := exec.Command("gcloud", "--project", app, "app", "deploy")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	cmd = exec.Command("gcloud", "--project", app, "app", "deploy", "cron.yaml")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func main() {
	apps := flag.String("apps", schedev, "What app to deploy to")
	instanceClass := flag.String("instance_class", "", fmt.Sprintf("Override the instance class setting for this app (%+v). Allowed values: F1, F2, F4", defaultInstanceClass))

	flag.Parse()

	if !validInstanceClasses[*instanceClass] {
		flag.Usage()
		os.Exit(1)
	}

	for _, app := range strings.Split(*apps, ",") {
		deployApp(app, *instanceClass)
	}
}
