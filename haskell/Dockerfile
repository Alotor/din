FROM din/base

RUN pacman -S --noconfirm ghc stack
RUN ln -s /usr/sbin/ghci /usr/local/bin/haskell

CMD ["/din"]
