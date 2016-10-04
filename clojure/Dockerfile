FROM din/base

RUN pacman -S --noconfirm jdk8-openjdk clojure wget
# RUN sudo -i -u work yaourt -S --noconfirm leiningen

RUN wget https://raw.githubusercontent.com/technomancy/leiningen/stable/bin/lein \
    -O /usr/local/bin/lein && \
    chmod 777 /usr/local/bin/lein

CMD "/din"
