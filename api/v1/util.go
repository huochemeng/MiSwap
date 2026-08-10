package v1

type chainIDMap map[int]string

var chainIDToChain = chainIDMap{
	1:        "eth",
	10:       "optimism",
	11155111: "sepolia",
}
