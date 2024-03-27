#!/bin/sh

NAME=captiveportal
CONFIG_DIR=/etc/$NAME

wget -O /bin/$NAME https://raw.githubusercontent.com/hub66899/captive-portal/master/captiveportal
chmod 777 /bin/$NAME

echo "fw4 set name: default allowed_mac"
read set_name
set_name="${set_name:-allowed_mac}"

mkdir -p $CONFIG_DIR

cat > /etc/init.d/$NAME << EOF
#!/bin/sh /etc/rc.common
START=99
USE_PROCD=1
start_service() {
    nft flush set inet fw4 $set_name

    procd_open_instance
    procd_set_param command "/bin/$NAME"
    procd_set_param env KEYWORDS_FILE="$CONFIG_DIR/keywords" DATA_FILE="$CONFIG_DIR/data"
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_close_instance
}
stop_service() {
    nft flush set inet fw4 $set_name
}
EOF
chmod +x /etc/init.d/$NAME
/etc/init.d/$NAME enable
/etc/init.d/$NAME start