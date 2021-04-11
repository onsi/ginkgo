package block_contest

func ReadTheChannel(c chan bool) {
	<-c
}

func SlowReadTheChannel(c chan bool) {
	<-c
}
