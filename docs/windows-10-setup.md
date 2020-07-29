# Windows 10 Linux Printer configuration

1. Open `Settings > Printers & Scanners`
1. Click `Add a printer or scanner`
1. Wait a little while for the `The printer that I want isn't listed` option and click it
1. Select `Add a local printer or network printer with manual settings` and click `Next`
1. Select `Create a new port` with type `Standard TCP/IP port` and click `Next`
1. Enter the host-name or IP address of the virtual printer, and leave the port name as default, and disable the option to query the printer, then click `Next`
1. Wait for the auto detection to complete
1. Select `Custom` and click then `Settings` button
1. Select protocol `LPR` and enter a `Queue name` of your choice e.g. `Alarm` then click `OK`
1. Click `Next`
1. Select manufacturer `Generic`, and Printer `Generic/Text Only` and click `Next`
1. Select to keep the existing driver and click `Next`
1. Enter a name for you new printer and click `Next`
1. Select `Do not share this printer` and click `Next`
1. Do not print a test page, and click `Finished`.

You virtual printer is ready to use. A quick test can be done using `notepad` typing a message and then clicking `File > Print` and selecting your new virtual printer.

