package testutils

func Contains(s []string, searchterm string) bool {
    for _, item := range s {
        if item == searchterm {
            return true
        }
    }
    return false
}