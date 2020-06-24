properties([gitLabConnection('Gitlab')])
node {
  checkout scm
  gitlabCommitStatus {
    docker.withRegistry('https://docker-push.vixns.net/', 'docker-push.vixns.net') {
            docker.build('vixns/acme.sh.ws').push()
    }
  }
}
