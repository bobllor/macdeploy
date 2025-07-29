package utils

func CheckError(err error) error {
	if err != nil {
		return err
	}

	return nil
}

// BuildString is used to create a variable name and its value.
func BuildString(varName string, varValue string) string {
	return varName + "=\"" + varValue + "\""
}
