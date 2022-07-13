# satellight

Light communication for light satellites travelling at almost light speed

```
 ______     ______     ______   ______     __         __         __     ______     __  __     ______  
/\  ___\   /\  __ \   /\__  _\ /\  ___\   /\ \       /\ \       /\ \   /\  ___\   /\ \_\ \   /\__  _\ 
\ \___  \  \ \  __ \  \/_/\ \/ \ \  __\   \ \ \____  \ \ \____  \ \ \  \ \ \__ \  \ \  __ \  \/_/\ \/ 
 \/\_____\  \ \_\ \_\    \ \_\  \ \_____\  \ \_____\  \ \_____\  \ \_\  \ \_____\  \ \_\ \_\    \ \_\ 
  \/_____/   \/_/\/_/     \/_/   \/_____/   \/_____/   \/_____/   \/_/   \/_____/   \/_/\/_/     \/_/
```

___

`satellight` is a small simulation of a fleet of individual satellites detecting and reporting ozone levels to a
ground-station.
It was created to showcase the fog-specific challenges of reliable message delivery when encountering churn in the
network.

Each component consists of a `sender`, a `receiver`, and some application logic.
While transmission of an individual message happens synchronously, i.e over HTTP with TCP,
communication between satellites and ground-station happen asynchronously, i.e not as an HTTP response.

## Satellites

Each satellite travels through space between random points.
At each location, an ozone level measurement is taken and transmitted to the groundstation.

If delivery to the ground station fails, for example due to bad weather conditions or space debris, messages are queued
for later retry.
Delivery is retried with a linear backoff, each failed delivery postpones the next retry by one more second.

Satellites can also crash or otherwise fail (e.g due to solar flares or high radiation levels) with a probability of 1%.

### Runtime Flags

```bash
Usage of satellites:
  -endpoint string
        Groundstation endpoint (default "http://localhost:8000")
  -satelliteCount int
        Count of satellites launched (default 5)
```

## Ground-station

When the ground-station receives a message containing a critically low ozone reading,
it determines the closest satellite based on the latest data it has and sends it to the location to fix it.

> In order to make the cases more interesting,
> we exclude the satellite that created the reading from the list of potential fixing satellites

Just like the satellites, the ground-station queues failed messages up for later delivery.
It also applies the linear backoff in the same way.

### Runtime Flags

```bash
Usage of groundstation:
  -groundstation string
        address to listen on (default ":8000")
  -satellites string
        Base URL of the satellites (default "http://localhost")
```

## Networking

We simulate network latency and jitter by artificially delaying each delivery by a random value between 1 and 5 seconds.

Each satellite starts their own receiver and listens on a fixed port: `9000 + <SatelliteID>`.
This enables the ground-station to send messages to each satellite asynchronously instead of as a response
to an incoming message.

