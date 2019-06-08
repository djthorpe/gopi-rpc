#!/bin/bash
# Script to install Gaffer on a Raspberry Pi
# Author: David Thorpe <djt@mutablelogic.com>
#
# Usage:
#   install-gaffer.sh [-f] [-u username]
#
# Flag -f will remove any existing installations first
# Flag -u will be used to determine which user gaffer runs under


#####################################################################

# PREFIX is the parent directory of the gaffer setup
PREFIX="/opt/gaffer"
# USERNAME is the username for the gaffer processes
USERNAME="gopi"
# VARPATH
VARPATH="/opt/gaffer/var"
# FORCE set to 1 will result in any existing installation being removed first
FORCE=0
# SSL
SSL_DAYS=99999
SSL_KEY="selfsigned.key"
SSL_CERT="selfsigned.cert"

# Repo folder
CURRENT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

#####################################################################
# PROCESS FLAGS

while getopts 'fu:' FLAG ; do
  case ${FLAG} in
    f)
	  FORCE=1
      ;;
    u)
	  USERNAME=${OPTARG}
      ;;      
    \?)
      echo "Invalid option: -${OPTARG}"
	  exit 1
      ;;
  esac
done

#####################################################################
# CHECKS

# Temporary location
TEMP_DIR=`mktemp -d`
if [ ! -d "${TEMP_DIR}" ]; then
  echo "Missing temporary directory: ${TEMP_DIR}"
  exit 1
fi

#####################################################################
# MAKE

GOBIN=${TEMP_DIR} make -f ${CURRENT_PATH}/../Makefile gaffer

#####################################################################
# INSTALL EXECUTABLES

echo "Installing gaffer under ${PREFIX}"

# Make prefix folder, install executables under bin
install -d "${PREFIX}" || exit -1
install -d "${PREFIX}/bin"
install ${TEMP_DIR}/gaffer-service "${PREFIX}/bin"
install ${TEMP_DIR}/gaffer-client "${PREFIX}/bin/gaffer"

# Install service executables under sbin
install -d ${PREFIX}/sbin
install ${TEMP_DIR}/helloworld-service "${PREFIX}/sbin" 

# Make etc and var
install -d ${PREFIX}/etc
sudo install -o ${USERNAME} -d ${PREFIX}/var

#####################################################################
# MAKE SSLKEY

if [ -f "${PREFIX}/etc/${SSL_CERT}" ] && [ -f "${PREFIX}/etc/${SSL_KEY}" ] && [ "${FORCE}" == "0" ] ; then
    echo "Not generating SSL key and cert, use -f otherwise"
else
  openssl req \
    -x509 -nodes \
    -newkey rsa:2048 \
    -keyout "${TEMP_DIR}/${SSL_KEY}" \
    -out "${TEMP_DIR}/${SSL_CERT}" \
    -days "${SSL_DAYS}" \
    -subj "/C=GB/L=London/O=mutablelogic/CN=mutablelogic.com"
  sudo install -o ${USERNAME} -m 0700 -D "${TEMP_DIR}/${SSL_KEY}" "${PREFIX}/etc"
  install -D "${TEMP_DIR}/${SSL_CERT}" "${PREFIX}/etc"
fi

#####################################################################
# MAKE USER

if ! id -u $USERNAME > /dev/null 2>&1 ; then 
    echo "Creating user $USERNAME"
    sudo useradd -d "${PREFIX}" -M -U -s /bin/false ${USERNAME}
fi

#####################################################################
# STOP & DISABLE SERVICE

SERVICE_NAME="gaffer"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
SERVICE_LOADED=`systemctl list-units | grep ${SERVICE_NAME}`

echo "Installing systemd service ${SERVICE_NAME}"
install "${CURRENT_PATH}/../etc/gaffer.service" "${PREFIX}/etc" 

if [ ! "${SERVICE_LOADED}" = "" ] ; then
  echo "Existing ${SERVICE_NAME} service loaded, stopping"
  sudo systemctl stop ${SERVICE_NAME} || exit -1
  sudo systemctl disable ${SERVICE_NAME} || exit -1
fi

if [ ! -e "${SERVICE_FILE}" ] ; then
    echo "Making symlink => ${SERVICE_FILE}"
    echo  ln -s "${PREFIX}/etc/gaffer.service" "${SERVICE_FILE}"
    sudo ln -s "${PREFIX}/etc/gaffer.service" "${SERVICE_FILE}"
fi

#####################################################################
# REPLACE SERVICE DETAILS

cat "${SERVICE_FILE}" \
  | sed "s/User=.*/User=${USERNAME}/g" \
  | sed "s/Group=.*/Group=${USERNAME}/g" \
  | sed "s/gaffer\.root\S*/gaffer\.root=${PREFIX//\//\\/}\/sbin/g" \
  | sed "s/gaffer\.path\S*/gaffer\.path=${PREFIX//\//\\/}\/var\/gaffer.json/g" \
  | sed "s/rpc\.sslcert\S*/rpc\.sslcert=${PREFIX//\//\\/}\/etc\/${SSL_CERT}/g" \
  | sed "s/rpc\.sslkey\S*/rpc\.sslkey=${PREFIX//\//\\/}\/etc\/${SSL_KEY}/g" \
  > ${SERVICE_FILE}

#####################################################################
# MAKE SYMLINK

# Enable service
echo "Enabling and reloading gaffer service"
sudo systemctl enable gaffer

