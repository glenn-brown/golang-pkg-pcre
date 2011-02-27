include $(GOROOT)/src/Make.inc

TARG=pcre

CGOFILES=\
	pcre.go

include $(GOROOT)/src/Make.pkg

.PHONY: install-debian
install-debian: _obj/$(TARG).a
	install -D _obj/$(TARG).a debian/golang-pkg-$(TARG)/$(pkgdir)/$(TARG).a 
