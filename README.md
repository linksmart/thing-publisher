## Synopsis

LinkSmart Thing Publisher (LSTP) is intended to continuously expose proprietary sensor data in the OGC SensorThing format.
The OGC SensorThing output is published over MQTT. The prioprietary sensor data is delivered via executables like scripts
or executable binaries. The exposing and formating of the data is done by the LSTP.

## Getting started

Edit the conf\thing-publisher.json and configure the MQTT broker URL. The default value is pointing to a local broker.
Execute the binary "thingpublisher". LSTP should be running now.
More documentation can be found here: https://docs.linksmart.eu/display/TP


## Hello World example

In the current example the mosquitto tools are used to show the funcionality of LSTP. The MQTT broker should be running
and configured. All commands are started from the directory where the thingpublisher executable was executed.

Type to subscribe to sensors topic: mosquitto_sub -t "LSTP/Datastreams(Temperature)/fixed sensor" -v
Deploy a LSTP Thing archive via: mosquitto_pub -t "LSTP/addthingarchive" -f ./TEST/agentarchives/temperature.tar.gz

After a while the subscribtion client will start producing output like:
LSTP/Datastreams(Temperature)/fixed sensor {"result":"20.0","phenomenonTime":"2017-10-23T12:41:54Z"}


## Contributors

check the vendoring directory for included libraries

## License

see LICENSE file
