package behavior

import petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"

// builtinDefinitions returns every supported Arcturus-compatible command.
func builtinDefinitions() []Definition {
	definitions := []Definition{
		mode(0, "free", petruntime.ActionClear), action(1, "sit", "sit"), action(2, "down", "lay"), mode(3, "here", petruntime.ActionHere),
		action(4, "beg", "beg"), action(5, "play dead", "ded"), mode(6, "stay", petruntime.ActionStay), mode(7, "follow", petruntime.ActionFollow),
		mode(8, "stand", petruntime.ActionClear), action(9, "jump", "jmp"), action(10, "speak", "spk"), action(11, "play", "pla"),
		mode(12, "silent", petruntime.ActionSilent), action(13, "nest", "slp"), need(14, "drink", petruntime.CommandNeedDrink), mode(15, "follow left", petruntime.ActionFollow),
		mode(16, "follow right", petruntime.ActionFollow), action(17, "play football", "kck"), mode(18, "come here", petruntime.ActionHere), action(19, "bounce", "jmp"),
		action(20, "flat", "lay"), action(21, "dance", "dan"), action(22, "spin", "trn"), action(23, "switch", "gst"),
		mode(24, "move forward", petruntime.ActionHere), action(25, "turn left", "trn"), action(26, "turn right", "trn"), action(27, "relax", "rlx"),
		action(28, "croak", "crk"), action(29, "dip", "dip"), action(30, "wave", "wav"), action(31, "mambo", "dan"),
		action(32, "high jump", "jmp"), action(33, "chicken dance", "dan"), action(34, "triple jump", "jmp"), action(35, "spread wings", "wng"),
		action(36, "breathe fire", "flm"), action(37, "hang", "rlx"), action(38, "torch", "eat"), action(40, "swing", "dan"),
		action(41, "roll", "trn"), action(42, "ring of fire", "flm"), need(43, "eat", petruntime.CommandNeedFood), action(44, "wag tail", "wag"),
		action(45, "count", "gst"), action(46, "breed", "gst"),
	}
	definitions[0].Aliases = []string{"libre"}
	definitions[1].Aliases = []string{"sientate", "siéntate"}
	definitions[2].Aliases = []string{"abajo"}
	definitions[3].Aliases = []string{"aqui", "aquí"}
	definitions[6].Aliases = []string{"quieto"}
	definitions[7].Aliases = []string{"sigueme", "sígueme"}
	return definitions
}
