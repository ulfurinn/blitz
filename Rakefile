def procs
  @procs ||= `grep -lr 'gen_proc:"gen_server"' src/bitbucket.org/ulfurinn/blitz/blizzard`.split(/\n/).map &:chomp
end

def gens
  @gens ||= `find src -name *.gen.go`.split(/\n/).map &:chomp
end

desc "Build"
task :default do
  sh "go install bitbucket.org/ulfurinn/gen_proc/gen_proc"
  procs.each do |pr|
    sh "bin/gen_proc #{pr}"
  end
  gens.each do |gen|
    sh "goimports -w=true #{gen}"
  end
  # sh 'stringer -type PGState src/bitbucket.org/ulfurinn/blitz/blizzard'
  sh 'go install -ldflags "-X main.patch `TZ=UTC date +%Y%m%d%H%M%S`" bitbucket.org/ulfurinn/blitz/cmd/...'
end

desc "Rebuild embedded"
task :rice do
  sh "rice -i bitbucket.org/ulfurinn/blitz/blizzard-lib embed-go"
  sh 'go install -ldflags "-X main.patch `TZ=UTC date +%Y%m%d%H%M%S`" bitbucket.org/ulfurinn/blitz/cmd/...'
end

desc "Rebuild dynamic"
task :norice do
  sh "rm src/bitbucket.org/ulfurinn/blitz/blizzard-lib/*.rice-box.go"
  sh 'go install -ldflags "-X main.patch `TZ=UTC date +%Y%m%d%H%M%S`" bitbucket.org/ulfurinn/blitz/cmd/...'
end
