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
        alertsTableStr += "\t<td>\n\t\t" + data[i]["PrintProgress"] + "%\n\t</td>\n";
        alertsTableStr += "\t<td>\n\t\t" + data[i]["PrintState"] + "\n\t</td>\n";
        alertsTableStr += "\t<td>\n\t\t" + data[i]["ReceiverName"] + "\n\t</td>\n";
        alertsTableStr += "\t<td>\n\t\t" + data[i]["ReceiverEmail"] + "\n\t</td>\n";
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