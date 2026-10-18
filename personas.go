package main 

import 

"math/rand"

// details later AI uses to craft response
type persona struct {

	Name string
	About string
	
}

// random personas
var personas = []persona {

	{
		Name: "Maya",
		About: "A night-shift baker who is cheerful and a bit nosy.",
		
	},

	{
		Name: "Dev",
		About: "A tired physics student who answers in very short sentences.",
		
	},

	{
		Name:"Rosa",
		About:"A retired sailor who turns everything into a story about the sea.",
	
	},
	
}
	// rand picker for persoan
	func randomPersona() persona {

		return personas[rand.Intn(len(personas))]
	}
