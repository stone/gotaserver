# GOTAserver

A Simple http server for OTA upgrades of ESP8266/Arduino.

The main difference between the GOTAserver than placing a firmware on a regular HTTP server
 is that it only serves firmware to the connecting device if there's a newer version of the
 firmware. If there are no firmware noting gets served to the connecting device.

Choose a directory, place your firmware files there, they need
to be in the following format format: /project/[anything]_<major>.<minor>.<suffix>
where [anything] can be an alpha numeric string describing firmware, 
like fermentrack, sonoff. <major> and <minor> are
ints describing the version of the firmware and the directory is the project.

***Examples:***

Directory layout:

    /srv/firmwaredir/esp8266-sonoff/esp8266_1.0.bin
    /srv/firmwaredir/esp8266-sonoff/esp8266_1.1.bin
    /srv/firmwaredir/fermentrack/firmware_2.2.bin
    /srv/firmwaredir/fermentrack/firmware_2.6.bin

Example requests for above files:

    /esp8266-sonoff/1.0/ Then esp8266_1.1.bin is served.
    /esp8266-sonoff/1.1/ 404 Not Found (no need for update already got latest version).
    /fermentrack/firmware_2.3.bin Then firmware_2.6.bin is served.
    /fermentrack/firmware_0.1.bin Then firmware_2.6.bin is served.
    /fermentrack/firmware_2.9.bin 404 Not Found (version later than our latest).


Configuration example:

    {
        "FirmwareDir": "/srv/firmwaredir",
        "FirmwareSuffix": "*.bin",
        "ServerHostPort": "localhost:8082"
    }

# Running GOTAserver

    $ gotaserver -h
    Usage of gotaserver:
        -c string
            Configuration file
        -d string
            Directory to serve firmware-files from (default "firmwares")
        -p string
            <host>:<port> to listen to (default "127.0.0.1:8000")
        -s string
            Suffix of files to use as firmware (default "*.bin")

    $ gotaserver -c cfg.json
    ============[ GOTAserver ]==================
    Firmware Directory  : /srv/firmwaredir
    Firmware Suffix     : *.bin
    Server listening on : localhost:8082
    ===========================================
    2017/04/18 20:21:42 PRJ: fermentrack REQ-VERSION: 6.1 FOUND:  firmware_6.2.bin
    2017/04/18 20:21:42 [GET] "/fermentrack/6.1/" 1.547301ms
    2017/04/18 20:21:48 PRJ: fermentrack REQ-VERSION: 1.0 FOUND:  firmware_6.2.bin
    2017/04/18 20:21:48 [GET] "/fermentrack/1.0/" 173.343µs
    2017/04/18 20:22:04 Already got the latest version
    2017/04/18 20:22:04 [GET] "/fermentrack/6.9/" 112.32µs
    2017/04/18 20:22:09 Already got the latest version
    2017/04/18 20:22:09 [GET] "/fermentrack/6.2/" 112.121µs
    2017/04/18 20:22:13 PRJ: fermentrack REQ-VERSION: 6.1 FOUND: firmware_6.2.bin
    2017/04/18 20:22:13 [GET] "/fermentrack/6.1/" 160.508µs


# Arduino setup

    #include <ArduinoOTA.h>
    #include <ESP8266HTTPClient.h>
    #include <ESP8266httpUpdate.h>
    #define OTA_SERVER 192.168.1.10:8080
    #define PROJECT_NAME fermentrack
    #define VERSION 1.1

    void doHTTPUpdate() {
        if (WiFi.status() != WL_CONNECTED) return;
        t_httpUpdate_return ret = ESPhttpUpdate.update("http://" OTA_SERVER "/" PROJECT_NAME "/", VERSION "/");
        switch(ret) {
            case HTTP_UPDATE_FAILED:
                Serial.printf("HTTP_UPDATE_FAILED Error (%d): %s\r\n", ESPhttpUpdate.getLastError(), ESPhttpUpdate.getLastErrorString().c_str());
                break;
            case HTTP_UPDATE_NO_UPDATES:
                Serial.println("HTTP_UPDATE_NO_UPDATES");
                break;
            case HTTP_UPDATE_OK:
                Serial.println("HTTP_UPDATE_OK");
                break;
        }
    }

ArduinoOTA will make a request to:
http://192.168.1.10:8080/fermentrack/1.1/

GOTAserver will then serve the latest version of fermentrack firmware
located in the directory firmwaredir/fermentrack/ and with our exampel above the file would be: firmware_2.6.bin. 
