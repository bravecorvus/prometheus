# Setting Up Raspberry Pi to work on Eduroam

### By Andrew Lee


## Introduction
This is a guide I put together in order to get my Raspberry Pi to work on my school's eduroam network.

Just to make things easier,
 ```
sudo su
 ```
(most of these commands will need root permissions)

The default cat_installer shell script file did not install the correct profiles and I needed to execute this file, use the outputs, and manually configure the configuration files. I am compiling this information for anyone who is having trouble connecting a single-board computer to the St. Olaf Eduroam Network specifically, or your eduroam network.


## First Cycle of Installation (Causing a failed script execution to generate the necessary configuration files)

Basically, this is splitting the process into two parts. 1st, we will purposefully cause a failed eduroam installation in order to generate the necessary .pem key file and the correct wpa_supplicant.

### Download The Shell Script
Download the shell commands from this link (or your equivalent cat_installer.sh scripts provided by your school): 
https://www.stolaf.edu/files/it/eduroam/eduroam-linux-SOC.sh

Go into that folder, and give the file execution permissions

```
chmod 777 eduroam-linux-SOC.sh
```

### Execute the Script (This will fail, but do not worry)

Execute the script.
The script will now execute. Enter your St. Olaf Login details (with full credintials [e.g. username@stolaf.edu])
The script will tell you it failed but do not worry, this was our aim.
Make sure you click yes on the options asking if you want the script to generate the wpa_supplicant file in plaintext format. (We will need this later)



### Backup the Generated Config Files

Now the failed runthrough of the script generated the wpa_supplicant profile that we will need later.

```
cd /root/.cat_installer/
```

You will see 2 files in there: ca.pem and cat_installer.conf. Copy these files into a directory of your choice that you will remember later. Basically, even the next step of successfully installing the eduroam profile will not allow you to connect to the internet since the script fails to change some key configuration files.

Create a new directory in root called .cat_installer, and copy the ca.pem key file in there.

```
cd /root
mkdir .cat_installer
cp /path-to-ca.pem /root/.cat_installer/ca.pem
```


Next, we will do the successful installation of the network profile for eduroam



## Second Installation Cycle (Successful)

### Installing the Necessary Software Components to Get a Successful Installation of Eduroam Profiles

Before doing anything, install the following utilities in your linux distribution (it will be used in the Shell ):

```
apt-get install dbus
apt-get install re
apt-get install uuid
apt-get install network-manager
```

### Modify NetworkManager.conf So it is the Default Controller for Wifi

```
nano /etc/NetworkManager/NetworkManager.conf

[main]
plugins=ifupdown.keyfile

[ifupdown]
managed=true
```

Where is sas "managed=false", replace with "managed=true"


### Execute the CAT_INSTALLER Shell Script Again (This Time, it Will Run Successfully)

Now go back and run eduroam-linux-SOC.sh
This time around, everything will execute correctly without a problem.

## Entering In The Correct Values in the wpa_supplicant.conf file

Now open cat_installer.conf (that you backed up as the last step of the failed configuration step) in a text editor

open /etc/wpa_supplicant/wpa_supplicant.conf in an editor of your choice.

```
nano /etc/wpa_supplicant/wpa_supplicant.conf
```

Leaving the top part as is

```
ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1
country=US
```

remove the rest of the various network configurations

copy and paste the information stored in cat_installer.conf into the wpa_supplicant.confhow t

Your final code should look something like this:

```
ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1
country=US

network={
        ssid="eduroam"
        key_mgmt=WPA-EAP
        pairwise=CCMP
        group=CCMP TKIP
        eap=PEAP
        identity="username@stolaf.edu"
        password="PassWord"
        ca_cert="/root/.cat_installer/ca.pem"
        domain_suffix_match="ad.stolaf.edu"
        phase2="auth=MSCHAPV2"
}
```

And save the information.

## Editing /etc/network/interfaces To Force NetworkManager to use the wpa_supplicant.conf As the Default Settings to Connect

Next, open /etc/network/interfaces on your favorite code editor.

```
nano /etc/network/interfaces
```

Now this part is necessary because we will be telling the current network handler NetworkManager, to get all the configuration information from the wpa_supplicant.config file we just patched to work for eduroam.

By default, there is alot of stuff enabled.
Now go through and comment out every line of the text by putting a "#" in front of each line. (we will comment out rather than delete just in case we need to reset everything)

At the bottom, paste the following code:

```
auto wlan0
allow-hotplug wlan0
iface wlan0 inet dhcp
        wpa-ssid eduroam
        pre-up wpa_supplicant -B -Dwext -i wlan0 -c/etc/w$
        post-down killall -q wpa_supplicant
```

## Restart Machine

Finally, restart the machine. At this point, you should see your machine as connected to eduroam

If you are still stuck with no luck with connecting to eduroam, here is when the trial and error comes in. Basically, you will be editing 3 files:

```
/etc/network/interfaces
/etc/NetworkManager/NetworkManager.conf
/etc/wpa_supplicant/wpa_supplicant.conf
```

So open them up using your favorite editor, then also have a spare Terminal open on the root user.

Basically, using the picture I have down below, edit the various file settings. (Note: You should always create backup copies of any files before you edit them so you can replace the new one if it really screws up your settings)

Every time you edit a file and save it, you will run the following command
```
systemctl daemon-reload
/etc/init.d/networking restart
```

## Fixing System Time

After you restart there is just one more step. Because the Pi does not come with a built in internal clock, the system date/time values will be so off that you will not be able to connect to the internet. (Which means you can't connect to an NTP server to update your time)

Therefore, manually change your time to the correct time. After this, you should be able to connect to the internet without an issue.

Don't forget to remove the eduroam-linux-SOC.sh. We gave it 777 permissions (out of laziness) so after you get connected, you definitely don't want to leave this file around.


## Here is my current setup for reference
![Correct Set Up](assets/Confi.png)
