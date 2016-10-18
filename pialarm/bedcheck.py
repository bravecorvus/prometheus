from Adafruit_ADS1x15 import ADS1x15

gain = 4096  # +/- 4.096V
sps = 250  # 250 samples per second
ADS1015 = 0x00  # 12-bit ADC
adc = ADS1x15(ic=ADS1015)

print(adc.readADCSingleEnded(0, gain, sps) / 1000)
