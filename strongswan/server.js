const http = require('http');

const requestListener = function (req, res) {
  res.writeHead(200);
  res.end('Hello from ' + process.env.VPN_LOCAL_PEER + "\n");
}

const server = http.createServer(requestListener);
server.listen(8080);

async function startTemporary(time) {
  const requestListener = function (req, res) {
    res.writeHead(200);
    res.end('Temporary hello from ' + process.env.VPN_LOCAL_PEER + "\n");
  }
  const server = http.createServer(requestListener);
  const startedServer = server.listen(8090);
  console.log("Started server")
  await new Promise(a => setTimeout(a, time))
  startedServer.close()
  console.log("Stoped server")
  await new Promise(a => setTimeout(a, time))
  return startTemporary(time)
}

startTemporary(5000)