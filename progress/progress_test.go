package progress

func ExampleDrawProgress() {
	DrawProgress("Test", 10, 100)
	// Output:Test	=====>                                             	10%		(10/100)
}

func ExampleDrawProgressTotal() {
	DrawProgress("Test", 100, 100)
	// Output:Test	==================================================>	100%		(100/100)
}
