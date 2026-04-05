package generator

const managedBy = "deployer"

func BuildLabels(name, partOf string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       name,
		"app.kubernetes.io/part-of":    partOf,
		"app.kubernetes.io/managed-by": managedBy,
	}
}
