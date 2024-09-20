let eIbcClientStarted = false;
let firstLoadEIbcClientStatus = true;

const reloadPartialBalances = () => {
    fetch('/partial/balances')
        .then(response => response.text())
        .then(html => {
            document.getElementById('partial-balances').innerHTML = html;
        });
}

const reloadPartialEIbcClientStatus = () => {
    fetch('/partial/eibc-client/logs')
        .then(response => response.text())
        .then(html => {
            document.getElementById('partial-eibc-client-logs').innerHTML = html;
        });

    postJson('update eIBC-client status', '/eibc-client/status', {}, (data) => {
        const inputDenom = $('#eibc-client-denom');
        const inputMinFee = $('#eibc-client-min-fee');

        eIbcClientStarted = data.result.running === true;
        if (firstLoadEIbcClientStatus) {
            $('#eibc-client-start').prop('disabled', false);
            inputDenom.prop('disabled', false);
            inputMinFee.prop('disabled', false);
            firstLoadEIbcClientStatus = false;
        }

        const conditionalReplace = (selector, newValue) => {
            if (newValue && newValue !== "") {
                const currentValue = selector.val();
                if (currentValue !== newValue) {
                    if (eIbcClientStarted) {
                        // force replace
                        selector.val(newValue);
                    } else if (currentValue === "") {
                        // replace only if empty
                        selector.val(newValue);
                    }
                }
            }
        }
        conditionalReplace(inputDenom, data.result.denom);
        conditionalReplace(inputMinFee, data.result.min_fee_percent);
        switchStateEIbcClientElements();
    }, (data) => {
        addError('failed to update eIBC client status', data);
    })
}

const toggleStartStopEIbcClient = () => {
    const btnStartStop = $('#eibc-client-start');
    btnStartStop.prop('disabled', true);
    setTimeout(() => {
        btnStartStop.prop('disabled', false);
    }, 1000);

    if (eIbcClientStarted) {
        postJson('stop eIBC client', '/eibc-client/stop', {}, (data) => {
            eIbcClientStarted = false;
            switchStateEIbcClientElements();
        }, (data) => {
            addError('failed to stop eIBC client:', data.result);
        })
    } else {
        postJson('start eIBC client', '/eibc-client/start', {
            denom: $('#eibc-client-denom').val(),
            min_fee_percent: $('#eibc-client-min-fee').val(),
        }, (data) => {
            eIbcClientStarted = true;
            switchStateEIbcClientElements();
        }, (data) => {
            addError('failed to start eIBC client:' + JSON.stringify(data.result));
        })
    }
}

const switchStateEIbcClientElements = () => {
    const started = eIbcClientStarted === true;
    const btnStart = $('#eibc-client-start');
    const inputDenom = $('#eibc-client-denom');
    const inputMinFee = $('#eibc-client-min-fee');
    if (started) {
        btnStart.addClass('btn-danger').removeClass('btn-primary');
        btnStart.val('Stop');
    } else {
        btnStart.addClass('btn-primary').removeClass('btn-danger');
        btnStart.val('Start eIBC client');
    }
    inputDenom.prop('disabled', started);
    inputMinFee.prop('disabled', started);
}

const handleApiResponse = (data, onSuccess, onError) => {
    if (data) {
        if (data.message === "OK") {
            onSuccess(data);
        } else if (data.message === "NOTOK") {
            if (onError) {
                onError(data);
            } else {
                addError('response data indicate error:' + JSON.stringify(data));
            }
        } else {
            addError('unknown response data:' + data.toString());
        }
    } else {
        addError('no response data');
    }
}

const addError = (message, err) => {
    if (err) {
        message = message + ': ' + err.toString();
    }

    const div = document.createElement('div');
    div.appendChild(document.createTextNode(message));

    const errorElement = document.getElementById('error-panel');
    errorElement.appendChild(div);
}

const postJson = function (actionName, url, data, onSuccess, onError) {
    $.ajax({
        type: "POST",
        url: url,
        data: JSON.stringify(data),
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function(data) {
            handleApiResponse(data, onSuccess, onError);
        },
        error: (xhr, status, error) => {
            let errMsg = `failed to [${actionName}]`;
            if (status) {
                errMsg += `, text status = '${status}'`;
            }
            if (xhr && xhr.responseText) {
                errMsg += `, response text = '${xhr.responseText}'`;
            }
            if (error) {
                errMsg += `, error = '${error}'`;
            }
            addError(errMsg);
        }
    })
}
