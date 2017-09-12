function updateAlertsTable() {
  $.ajax({
    type: 'GET',
    url: '/alerts',
    timeout: 2000,
    success: function(data) {
      if (data == null) {
        return;
      }
      console.log(JSON.stringify(data))
      var alertsTableStr = "";
      for (var i = 0; i < data.length; i++) {
        alertsTableStr += "<tr>\n";
        alertsTableStr += "\t<td>\n\t\t" + data[i]["PrinterSerial"] + "\n\t</td>\n";
        alertsTableStr += "\t<td>\n\t\t" + data[i]["PrintName"] + "\n\t</td>\n";
        var printProgress = data[i]["PrintProgress"].toFixed(1);
        // alertsTableStr += "\t<td>\n\t\t" + printProgress + "%\n\t</td>\n";
        alertsTableStr += '\t<td>\n\t\t<div class="progress">\n<div class="progress-bar" role="progressbar" style="width: ' +
          printProgress + '%;" aria-valuenow="' + printProgress + '" aria-valuemin="0" aria-valuemax="100">' +
          printProgress + '%</div>\n</div>\n\t</td>\n';
        alertsTableStr += "\t<td>\n\t\t" + data[i]["PrintState"] + "\n\t</td>\n";
        alertsTableStr += "\t<td>\n\t\t" + data[i]["ReceiverName"] + "\n\t</td>\n";
        alertsTableStr += "\t<td>\n\t\t" + data[i]["ReceiverEmail"] + "\n\t</td>\n";
        var alertStatus = "";
        if (printProgress == 0 && data[i]["PrintState"] == "") {
          alertsTableStr += "\t<td class=\"text-warning\">\n\t\tLoading\n\t</td>\n";
        } else if (!data[i]["ShouldEmail"]) {
          alertsTableStr += "\t<td class=\"text-danger\">\n\t\tWon't send\n\t</td>\n";
        } else if (data[i]["SentEmail"]) {
          alertsTableStr += "\t<td class=\"text-success\">\n\t\tSent\n\t</td>\n";
        } else {
          alertsTableStr += "\t<td class=\"text-primary\">\n\t\tPending\n\t</td>\n";
        }
        alertsTableStr += "</tr>\n";
      }
      $("#alerts_body").html(alertsTableStr);
      window.setTimeout(updateAlertsTable, 10000);
    },
    error: function (XMLHttpRequest, textStatus, errorThrown) {
      console.log(errorThrown);
      window.setTimeout(updateAlertsTable, 60000);
    }
  });
}

$(document).ready(function() {
  var INTERVAL = 10000;
  setInterval(updateAlertsTable(), INTERVAL);
});