desc "Build"
task :default do
  sh "bin/gen_proc src/bitbucket.org/ulfurinn/blitz/blizzard-lib/blizzard.go"
  sh "bin/gen_proc src/bitbucket.org/ulfurinn/blitz/blizzard-lib/proc_group.go"
  sh "bin/gen_proc src/bitbucket.org/ulfurinn/blitz/blizzard-lib/proc.go"
  sh "goimports -w=true src/bitbucket.org/ulfurinn/blitz/blizzard-lib/blizzard.gen.go"
  sh "goimports -w=true src/bitbucket.org/ulfurinn/blitz/blizzard-lib/proc_group.gen.go"
  sh "goimports -w=true src/bitbucket.org/ulfurinn/blitz/blizzard-lib/proc.gen.go"
  sh 'go install -ldflags "-X main.patch `date +%s`" bitbucket.org/ulfurinn/blitz/...'
end
