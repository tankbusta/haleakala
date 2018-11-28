BUILDREV		:= $(shell git log --pretty=format:'%h' -n 1)
BUILDDATE		:= $(shell date +%s)

ifdef RELEASE
	LDSTRIP		+= -w -s
endif

LDFLAGS			+=	-X haleakala.CurrentVersion=${BUILDREV} -X haleakala.BuildTime=${BUILDDATE}

ifdef STATIC_BUILD
	CGO_LDFLAGS	+=	'-lstdc++ -lm'
	LDFLAGS		+= 	-extldflags "-static"
	GO_BUILD_FLAGS	+=	-a
endif

build:
	CGO_LDFLAGS=$(CGO_LDFLAGS) go build $(GO_BUILD_FLAGS) -o haleakala -tags "$(TAGS)" --ldflags "$(LDFLAGS)" cmd/haleakala.go

clean:
	rm -f haleakala

