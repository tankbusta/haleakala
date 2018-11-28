package sounds

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type SoundCollections []*SoundCollection

func (s *SoundCollections) GetFullList() map[string][]string {
	out := make(map[string][]string)
	for _, collection := range *s {
		logrus.Debugf("Loading collection %s\n", collection.Prefix)
		out[collection.Prefix] = make([]string, len(collection.Sounds))
		for i, sound := range collection.Sounds {
			out[collection.Prefix][i] = sound.Name
		}
	}
	return out
}

type SoundCollection struct {
	Prefix    string
	Commands  []string
	Sounds    []*Sound
	ChainWith *SoundCollection

	soundRange int
}

func (sc *SoundCollection) Load(basePath string) {
	logrus.Debugf("Loading collection %s with %d sounds", sc.Prefix, len(sc.Sounds))
	for _, sound := range sc.Sounds {
		sc.soundRange += sound.Weight
		if err := sound.Load(basePath, sc); err != nil {
			logrus.WithFields(logrus.Fields{
				"err":   err,
				"sound": fmt.Sprintf("%s_%s", sc.Prefix, sound.Name),
			}).Error("Failed to load sound")
		}
	}
}

func (s *SoundCollection) Random() *Sound {
	var (
		i      int
		number int = randomRange(0, s.soundRange)
	)

	for _, sound := range s.Sounds {
		i += sound.Weight

		if number < i {
			return sound
		}
	}
	return nil
}

var airhorn = &SoundCollection{
	Prefix: "airhorn",
	Commands: []string{
		"airhorn",
	},
	Sounds: []*Sound{
		createSound("default", 1000, 250),
		createSound("reverb", 800, 250),
		createSound("spam", 800, 0),
		createSound("tripletap", 800, 250),
		createSound("fourtap", 800, 250),
		createSound("distant", 500, 250),
		createSound("echo", 500, 250),
		createSound("clownfull", 250, 250),
		createSound("clownshort", 250, 250),
		createSound("clownspam", 250, 0),
		createSound("highfartlong", 200, 250),
		createSound("highfartshort", 200, 250),
		createSound("midshort", 100, 250),
		createSound("truck", 10, 250),
	},
}

var khaled = &SoundCollection{
	Prefix:    "another",
	ChainWith: airhorn,
	Commands: []string{
		"anotha",
		"anothaone",
	},
	Sounds: []*Sound{
		createSound("one", 1, 250),
		createSound("one_classic", 1, 250),
		createSound("one_echo", 1, 250),
	},
}

var cena = &SoundCollection{
	Prefix: "jc",
	Commands: []string{
		"johncena",
		"cena",
	},
	Sounds: []*Sound{
		createSound("airhorn", 1, 250),
		createSound("echo", 1, 250),
		createSound("full", 1, 250),
		createSound("jc", 1, 250),
		createSound("nameis", 1, 250),
		createSound("spam", 1, 250),
	},
}

var ethan = &SoundCollection{
	Prefix: "ethan",
	Commands: []string{
		"ethan",
		"eb",
		"ethanbradberry",
		"h3h3",
	},
	Sounds: []*Sound{
		createSound("areyou_classic", 100, 250),
		createSound("areyou_condensed", 100, 250),
		createSound("areyou_crazy", 100, 250),
		createSound("areyou_ethan", 100, 250),
		createSound("classic", 100, 250),
		createSound("echo", 100, 250),
		createSound("high", 100, 250),
		createSound("slowandlow", 100, 250),
		createSound("cuts", 30, 250),
		createSound("beat", 30, 250),
		createSound("sodiepop", 1, 250),
	},
}

var cow = &SoundCollection{
	Prefix: "cow",
	Commands: []string{
		"stan",
	},
	Sounds: []*Sound{
		createSound("herd", 10, 250),
		createSound("moo", 10, 250),
		createSound("x3", 1, 250),
	},
}

var birthday = &SoundCollection{
	Prefix: "birthday",
	Commands: []string{
		"birthday",
		"bday",
	},
	Sounds: []*Sound{
		createSound("horn", 50, 250),
		createSound("horn3", 30, 250),
		createSound("sadhorn", 25, 250),
		createSound("weakhorn", 25, 250),
	},
}

var wow = &SoundCollection{
	Prefix: "wow",
	Commands: []string{
		"wowthatscool",
		"wtc",
	},
	Sounds: []*Sound{
		createSound("thatscool", 50, 250),
	},
}

var vince = &SoundCollection{
	Prefix: "vince",
	Commands: []string{
		"vince",
	},
	Sounds: []*Sound{
		createSound("dick", 25, 250),
		createSound("fuckthis", 25, 250),
		createSound("dontcare", 30, 250),
		createSound("holdstill", 30, 250),
		createSound("knobhead", 30, 250),
		createSound("supertitans", 20, 250),
	},
}

var trump = &SoundCollection{
	Prefix: "trump",
	Commands: []string{
		"maga",
		"trump",
	},
	Sounds: []*Sound{
		createSound("wrong", 10, 250),
		createSound("hombres", 20, 250),
		createSound("wall", 25, 250),
		createSound("getitout", 25, 250),
		createSound("bing", 30, 250),
		createSound("mess", 30, 250),
		createSound("tractor", 40, 250),
		createSound("worstpres", 50, 250),
		createSound("3x", 50, 250),
		createSound("lovechina", 55, 250),
		createSound("mexican", 60, 250),
	},
}

var duke = &SoundCollection{
	Prefix: "duke",
	Commands: []string{
		"duke",
		"theking",
	},
	Sounds: []*Sound{
		createSound("bigdick", 10, 250),
		createSound("bigguns", 20, 250),
		createSound("birthcontrol", 25, 250),
		createSound("blowitoutyourass", 25, 250),
		createSound("yippiekaiay", 30, 250),
		createSound("yourface", 30, 250),
	},
}

var COLLECTIONS = SoundCollections{
	airhorn,
	khaled,
	cena,
	ethan,
	cow,
	birthday,
	wow,
	vince,
	trump,
}
