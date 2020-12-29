#export GOPATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

######################

#SOURCES_BASE = src/types/types.go src/readline/readline.go \
#	       src/reader/reader.go src/printer/printer.go \
#	       src/env/env.go src/core/core.go

######################

#SRCS = step0_repl.go step1_read_print.go step2_eval.go step3_env.go \
#       step4_if_fn_do.go step5_tco.go step6_file.go step7_quote.go \
#       step8_macros.go step9_try.go stepA_mal.go
#BINS = $(SRCS:%.go=%)

######################

#all: $(BINS)

#dist: mal

#mal: $(word $(words $(BINS)),$(BINS))
#	cp $< $@

#define dep_template
#$(1): $(SOURCES_BASE) src/$(1)/$(1).go
#	go build $$@
#endef

#$(foreach b,$(BINS),$(eval $(call dep_template,$(b))))

#clean:
#	rm -f $(BINS) mal

all:
	go build src/step0_repl/*.go
	go build -o step1_read_print src/step1_read_print/*.go
	go build src/step2_eval/*.go
	go build src/step3_env/*.go
	go build src/step4_if_fn_do/*.go
	go build src/step5_tco/step5_tco.go
	go build src/step6_file/step6_file.go
