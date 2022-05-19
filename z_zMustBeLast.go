package jar

func init() {
	// This is here to ensure that after the last alphabetical init() is called,
	// we are sure to update the Conf
	ConfigAdditions.Set(Conf)
}
