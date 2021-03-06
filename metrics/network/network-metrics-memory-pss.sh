#!/bin/bash

#  Copyright (C) 2017 Intel Corporation
#
#  This program is free software; you can redistribute it and/or
#  modify it under the terms of the GNU General Public License
#  as published by the Free Software Foundation; either version 2
#  of the License, or (at your option) any later version.
#
#  This program is distributed in the hope that it will be useful,
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  GNU General Public License for more details.
#
#  You should have received a copy of the GNU General Public License
#  along with this program; if not, write to the Free Software
#  Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
#
# Description:
#  Measures Proportional Set Size memory while running an 
#  inter (docker<->docker) network bandwidth using iperf2

SCRIPT_PATH=$(dirname "$(readlink -f "$0")")

source "${SCRIPT_PATH}/lib/network-test-common.bash"

# Set QEMU_PATH unless it's already set
QEMU_PATH=${QEMU_PATH:-$(get_qemu_path)}

# This script will perform all the measurements using a local setup

# Measures PSS memory while running bandwidth measurements
# using iperf2

function pss_memory {
	# Port number where the server will run
	local port=5001:5001
	# Using this image as iperf is not working
	# see (https://github.com/01org/cc-oci-runtime/issues/152)
	# Image name
	local image=gabyct/network
	# Total measurement time (seconds)
	# This is required in order to reduce standard deviation
	local total_time=10
	# This time (seconds) is required when
	# server and client are more stable, we need to
	# have server and client running for sometime and we
	# need to avoid to measure at the beginning of the running
	local middle_time=5
	# Name of the containers
	local server_name="network-server"
	local client_name="network-client"
	# Arguments to run the client
	local extra_args="-d"

	setup
	local server_command="iperf -p ${port} -s"
	local server_address=$(start_server "$server_name" "$image" "$server_command")

	local client_command="iperf -c ${server_address} -t ${total_time}"
	start_client "$extra_args" "$client_name" "$image" "$client_command" > /dev/null

	# Measurement after client and server are more stable
	echo >&2 "WARNING: sleeping for $middle_time seconds in order to have server and client stable"
	sleep ${middle_time}
	local memory_command="smem --no-header -c pss"
	${memory_command} -P ${QEMU_PATH} > "$result"

	local total_pss_memory=$(awk '{ total += $1 } END { print total/NR }' "$result")
	echo "The PSS memory is : $total_pss_memory Kb"

	save_results "network metrics memory" "pss" "$total_pss_memory" "kb"

	clean_environment "$server_name"
	$DOCKER_EXE rm -f ${client_name} > /dev/null
}

pss_memory
