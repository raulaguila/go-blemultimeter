# Bluetooth communication protocol

Tested with [AOPUTTRIVER AP-90EPD Digital Multimeter](https://www.amazon.ca/AOPUTTRIVER-Auto-ranging-Resistance-Capacitance-Temperature/dp/B07PB6XKJW)

Works with FS9721-LP3 based Bluetooth Multimeters, including:

* AOPUTTRIVER AP-90EPD
* Infurider YF-90EPD
* HoldPeak HP-90EPD

![](/screenshot/fs9721lp3.png)


* ### Example message:

    * Receives two slices of byte with eight and six bytes respectively.
    * When receive the slice with 8 bytes, save the slice.
    * When receive the slice with 6 bytes, concatenate with the saved slice and process.
    * Convert each byte to 8 digits binary.
    * Only the last four digits of each converted byte are important.
    * Concatenate all last 4 digits of each converted byte into a string of size (6 + 8) * 4 = 56.
    * Each character in this string represents an element on the LCD multimeter as shown in the image:

        ![](/screenshot/fs9721lp3_LCD.jpg)

