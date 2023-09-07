# Note: Intended to be run as "make run-blackbox-tests" or "make run-blackbox-ci"
#       Makefile target installs & checks all necessary tooling
#       Extra tools that are not covered in Makefile target needs to be added in verify_prerequisites()

load helpers_zot

function verify_prerequisites {
    return 0
}

function setup_file() {
    # Verify prerequisites are available
    if ! $(verify_prerequisites); then
        exit 1
    fi

    # Download test data to folder common for the entire suite, not just this file
    skopeo --insecure-policy copy --format=oci docker://ghcr.io/project-zot/golang:1.20 oci:${TEST_DATA_DIR}/golang:1.20
    # Setup zot server
    local zot_root_dir=${BATS_FILE_TMPDIR}/zot
    local zot_config_file=${BATS_FILE_TMPDIR}/zot_config.json
    local oci_data_dir=${BATS_FILE_TMPDIR}/oci
    local htpasswordFile=${BATS_FILE_TMPDIR}/htpasswd
    mkdir -p ${zot_root_dir}
    mkdir -p ${oci_data_dir}
    echo 'test:$2a$10$EIIoeCnvsIDAJeDL4T1sEOnL2fWOvsq7ACZbs3RT40BBBXg.Ih7V.' >> ${htpasswordFile}
    cat > ${zot_config_file}<<EOF
{
    "distSpecVersion": "1.1.0-dev",
    "storage": {
        "rootDirectory": "${zot_root_dir}"
    },
    "http": {
        "address": "127.0.0.1",
        "port": "8080",
        "auth": {
            "htpasswd": {
                "path": "${htpasswordFile}"
            }
        },
        "accessControl": {
            "repositories": {
                "**": {
                    "anonymousPolicy": ["read"],
                    "policies": [
                        {
                            "users": [
                                "test"
                            ],
                            "actions": [
                                "read",
                                "create",
                                "update"
                            ]
                        }
                    ]
                }
            }
        }
    },
    "log": {
        "level": "debug",
        "output": "${BATS_FILE_TMPDIR}/zot.log"
    }
}
EOF
    zot_serve ${ZOT_PATH} ${zot_config_file}
    wait_zot_reachable 8080
}

function teardown() {
    # conditionally printing on failure is possible from teardown but not from from teardown_file
    cat ${BATS_FILE_TMPDIR}/zot.log
}

function teardown_file() {
    zot_stop_all
}

@test "push image user policy" {
    run skopeo --insecure-policy copy --dest-creds test:test --dest-tls-verify=false \
        oci:${TEST_DATA_DIR}/golang:1.20 \
        docker://127.0.0.1:8080/golang:1.20
    [ "$status" -eq 0 ]
}

@test "pull image anonymous policy" {
    local oci_data_dir=${BATS_FILE_TMPDIR}/oci
    run skopeo --insecure-policy copy --src-tls-verify=false \
        docker://127.0.0.1:8080/golang:1.20 \
        oci:${oci_data_dir}/golang:1.20
    [ "$status" -eq 0 ]
}

@test "push image anonymous policy" {
    run skopeo --insecure-policy copy --dest-tls-verify=false \
        oci:${TEST_DATA_DIR}/golang:1.20 \
        docker://127.0.0.1:8080/golang:1.20
    [ "$status" -eq 1 ]
}
