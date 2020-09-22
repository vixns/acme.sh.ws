FROM vixns/acme.sh.ws
RUN adduser -S -u 987 -g 987 runner \
&& mv /root/.acme.sh /home/runner/.acme.sh \
&& ln -sf /home/runner/.acme.sh/acme.sh /usr/local/bin/acme.sh \
&& chown -R runner /home/runner

ENV HOME=/home/runner

USER runner
