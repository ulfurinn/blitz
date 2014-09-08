desc "Build"
task :default do
  sh "bin/gen_proc src/bitbucket.org/ulfurinn/blitz/blizzard-lib/blizzard.go"
  sh "bin/gen_proc src/bitbucket.org/ulfurinn/blitz/blizzard-lib/proc_group.go"
  sh "bin/gen_proc src/bitbucket.org/ulfurinn/blitz/blizzard-lib/proc.go"
  sh "bin/gen_proc src/bitbucket.org/ulfurinn/blitz/blizzard-lib/admin.go"
  sh "goimports -w=true src/bitbucket.org/ulfurinn/blitz/blizzard-lib/blizzard.gen.go"
  sh "goimports -w=true src/bitbucket.org/ulfurinn/blitz/blizzard-lib/proc_group.gen.go"
  sh "goimports -w=true src/bitbucket.org/ulfurinn/blitz/blizzard-lib/proc.gen.go"
  sh "goimports -w=true src/bitbucket.org/ulfurinn/blitz/blizzard-lib/admin.gen.go"
  sh 'go install -ldflags "-X main.patch `TZ=UTC date +%Y%m%d%H%M%S`" bitbucket.org/ulfurinn/blitz/...'
end

desc "Rebuild embedded"
task :rice do
  sh "rice -i bitbucket.org/ulfurinn/blitz/blizzard-lib embed-go"
  sh 'go install -ldflags "-X main.patch `TZ=UTC date +%Y%m%d%H%M%S`" bitbucket.org/ulfurinn/blitz/...'
end

desc "Rebuild dynamic"
task :norice do
  sh "rm src/bitbucket.org/ulfurinn/blitz/blizzard-lib/*.rice-box.go"
  sh 'go install -ldflags "-X main.patch `TZ=UTC date +%Y%m%d%H%M%S`" bitbucket.org/ulfurinn/blitz/...'
end
