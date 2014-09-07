app = proc do |env|
  body = env.to_s
  [
    200,          # Status code
    {             # Response headers
      'Content-Type' => 'text/html',
      'Content-Length' => body.length.to_s,
    },
    [body]        # Response body
  ]
end

run app
