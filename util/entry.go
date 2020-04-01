package util

func Execute() {
	// Load config
	loadConfig()

	// Init Redis Client
	initRedisClient()

	// Synchronize composer.phar
	go composerPhar("composerPhar", 1)

	// Synchronize packages.json
	go packagesJsonFile("PackagesJson", 1)

	// Update status
	go status("Status", 1)

	Wg.Add(1)

	for i := 0; i < 12; i++ {
		go providers("Provider", i)
	}

	for i := 0; i < 30; i++ {
		go packages("Packages", i)
	}

	for i := 0; i < 50; i++ {
		go dists("Dists", i)
	}

	for i := 0; i < 1; i++ {
		go distsRetry(403, i)
	}

	for i := 0; i < 1; i++ {
		go distsRetry(500, i)
	}

	for i := 0; i < 1; i++ {
		go distsRetry(502, i)
	}

}
