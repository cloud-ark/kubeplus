function validateData() {
  return true;
}

function get_all_resources(resource){
  //alert("Hello World");
  displayString = "1. abc-org-tenant1 <br> 2. abc-org-tenant2"

  num_of_instances = document.getElementById("num_of_instances");
  num_of_instances.innerHTML = "Calculating...";
  num_of_instances.setAttribute("style", "display:block;");

  cpu = "-";
  memory = "-";
  storage = "-";
  nw_ingress = "-";
  nw_egress = "-";
  set_metrics(cpu, memory, storage, nw_ingress, nw_egress);

  var xhttp;
  xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
      //Reference: https://stackoverflow.com/questions/33914245/how-to-replace-the-entire-html-webpage-with-ajax-response
        $("html").html(xhttp.responseText);
//        render_resources(this, resource);
    }
  };
  //url = "/getAll?resource=" + resource;
  url = "/service/" + resource
  xhttp.open("GET", url, true);
  xhttp.send();
}

function render_resources(xmlhttp, resource)
{
  data = xmlhttp.responseText;
  console.log(data);
  console.log("-------");
  data1 = JSON.parse(data)
  instances = data1[resource];
  console.log(instances)
  displayString = "";
  count = 1
  for (const val of instances) {
    displayString = displayString + count + ".&nbsp;&nbsp;" + val['name'] + "&nbsp;&nbsp;" + val['namespace'] + "<br>"
    count = count + 1
  }
  element = document.getElementById("service_information_space");
  element.innerHTML = displayString;
}

function get_resource_api_doc(resource){
  displayString = "Here is information about: " + resource
  //alert(displayString);
  var xhttp;
  xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
        render_resource_api_doc(this, resource);
    }
  };
  url = "/get_resource_manpage?resource=" + resource;
  xhttp.open("GET", url, true);
  xhttp.send();
}

function render_resource_api_doc(xmlhttp, resource)
{
  data = xmlhttp.responseText;
  console.log(data);
  console.log("-------");
  data1 = JSON.parse(data)
  manPage = data1[resource];
  console.log(manPage)

  //document.getElementById("input-form").style.display = "none";
  //document.getElementById("metrics-details").style.display = "none";

  elementsToHide = ["input-form", "metrics-details","consumption_string_id", "num_of_instances","create-status"];
  hideElements(elementsToHide);

  element = document.getElementById("man-page");
  element.innerHTML = manPage;
  document.getElementById("man-page").style.display = "block";
}

function get_resource(res_string) {

  elementsToHide = ["input-form", "man-page","create-status"];
  hideElements(elementsToHide);

  document.getElementById("metrics-details").style.display = "block";
  document.getElementById("metrics-details").setAttribute("class","table table-condensed table-striped table-bordered");
  document.getElementById("metrics-details").setAttribute("width","100%");

  cpu = "-";
  memory = "-";
  storage = "-";
  nw_ingress = "-";
  nw_egress = "-";
  set_metrics(cpu, memory, storage, nw_ingress, nw_egress);

  num_of_instances = document.getElementById("num_of_instances");
  num_of_instances.innerHTML = "Calculating...";
  num_of_instances.setAttribute("style", "display:block;");

  myArr = res_string.split(",")

  console.log("Name:" + myArr[0]);
  instance = myArr[0];
  namespace = myArr[1];
  resource = myArr[2];
  element = document.getElementById("consumption_string_id");
  element.innerHTML = "Details for " + instance;
  element.style.display = "block";

  // TODO: Make Ajax call to get metrics for <instance, namespace, resource>

  url = "/service/instance_data?resource=" + resource + "&instance=" + instance + "&namespace=" + namespace;
  var xhr = new XMLHttpRequest();
  xhr.open('GET', url, true);
  xhr.onreadystatechange = function () {
    if (this.readyState == 4 && this.status == 200) {
        fieldData = this.responseText;
        console.log(fieldData);
        console.log("-------");
        data1 = JSON.parse(fieldData)

        cpu = data1['cpu'];
        memory = data1['memory'];
        storage = data1['storage'];
        nw_ingress = data1['nw_ingress'];
        nw_egress = data1['nw_egress'];
        set_metrics(cpu, memory, storage, nw_ingress, nw_egress);

        app_url = data1['app_url']
        console.log("App URL:" + app_url)
        /*
        // Connections O/P is not relevant for Consumer so no need to display it.
        connections_op = data1['connections_op'];
        document.getElementById("connections_op").innerHTML = connections_op;
        document.getElementById("connections_op").style.display = "block";
        */

        app_url1 = "<hr><a href=\"" + app_url + "\">Application URL</a><hr>"
        document.getElementById("app_url").innerHTML = app_url1;
        document.getElementById("app_url").style.display = "block";

        log_data = data1['logs'];
        console.log(log_data)

        textarea = "<label style=\"font-size:large;\">Application Logs</label><br><p><textarea style=\"overflow:scroll;width:600px;height:200px\">" + log_data + "</p></textarea>";
        document.getElementById("app_logs_data").innerHTML = textarea;
        document.getElementById("app_logs_data").style.display = "block";

	deleteButton = document.createElement("Button");
	deleteButton.innerHTML = "Delete"
	deleteButton.setAttribute("onclick", "delete_instance('" + resource + "','" + instance + "','" + namespace + "')");
	document.getElementById("app_delete").appendChild(deleteButton);
        document.getElementById("app_delete").style.display = "block";

        /*
        logs_url = "<a href=\"" + "\">Application Logs</a><hr><br></br>"
        document.getElementById("app_logs_url").innerHTML = logs_url;
        document.getElementById("app_logs_url").style.display = "block";
        document.getElementById("app_logs_url").onclick = get_logs(resource, instance, namespace)
        */

        elementsToHide = ["num_of_instances"];
        hideElements(elementsToHide);
      }
    };
  xhr.send();
}

function delete_instance(resource, instance, namespace) {
  resource = resource.trim();
  instance = instance.trim();
  namespace = namespace.trim();
  url = "/service/instance_delete?resource=" + resource + "&instance=" + instance + "&namespace=" + namespace;
  var xhr = new XMLHttpRequest();
  xhr.open('GET', url, true);
  xhr.onreadystatechange = function () {
    if (this.readyState == 4 && this.status == 200) {
        fieldData = this.responseText;
        console.log(fieldData);
        console.log("-------");
        data1 = JSON.parse(fieldData);
        delete_status = data1['status'];
        console.log(delete_status);
	document.getElementById("consumption_string_id").innerHTML = delete_status;
	elementsToHide = ["app_url", "app_logs_data","app_delete","metrics-details"];
	hideElements(elementsToHide);
	nodeToDelete = document.getElementById(resource + "-" + instance);
	nodeToDelete.parentNode.removeChild(nodeToDelete);
    }
  }
  xhr.send();
}

function get_logs(resource, instance, namespace) {

  url = "/service/instance_logs?resource=" + resource + "&instance=" + instance + "&namespace=" + namespace;
  var xhr = new XMLHttpRequest();
  xhr.open('GET', url, true);
  xhr.onreadystatechange = function () {
    if (this.readyState == 4 && this.status == 200) {
        fieldData = this.responseText;
        console.log(fieldData);
        console.log("-------");
        data1 = JSON.parse(fieldData);
        log_data = data1['logs'];
        console.log(log_data)

        textarea = "<textarea style=\"overflow:scroll;width:300px;height:200px\">" + log_data + "</textarea>";
        document.getElementById("app_logs_data").innerHTML = textarea;
        document.getElementById("app_logs_data").style.display = "block";
      }
    };
  xhr.send();
}

function set_metrics(cpu, memory, storage, ingress, egress) {
  cpuelement = document.getElementById("total_cpu");
  cpuelement.innerHTML = cpu + "<br> (millicores)";

  memelement = document.getElementById("total_memory");
  memelement.innerHTML = memory + "<br> (mebibytes)" ;

  storageelement = document.getElementById("total_storage");
  storageelement.innerHTML = storage + "<br> (Giga bytes)" ;

  ingresselement = document.getElementById("total_nw_ingress");
  ingresselement.innerHTML = nw_ingress + "<br> (bytes)" ;

  egresselement = document.getElementById("total_nw_egress");
  egresselement.innerHTML = nw_egress + "<br> (bytes)" ;
}

function create_resource(resource) {

  document.getElementById("input-form").style.display = "block";

  elementsToHide = ["metrics-details","man-page","consumption_string_id","num_of_instances","create-status"];
  hideElements(elementsToHide);

 queryParam = "crd=" + resource;
 url = "/resourcespec?" + queryParam;
  console.log("URL:" + url);
  var xhttp = new XMLHttpRequest();
  xhttp.open("GET", url, true);

  xhttp.onreadystatechange = function () {
    if (this.readyState == 4 && this.status == 200) {
      fieldData = JSON.parse(this.responseText);
      console.log(fieldData);
      console.log("-------");
      fields = fieldData['resource_spec'];
      console.log(fields);

      formDetailsElement = document.getElementById("form-details");
      header = document.createElement("h4");
      header.innerHTML = "Enter data";
      formDetailsElement.appendChild(header);

      instanceSpec = document.createElement("textarea");
      instanceSpec.setAttribute("overflow", "scroll");
      instanceSpec.setAttribute("style", "height:95%; width:95%; margin-left:3%; margin-right:5%");
      instanceSpec.setAttribute("white-space","pre");
      instanceSpec.setAttribute("id","instanceSpec");
      instanceSpec.setAttribute("name","instanceSpec");
      instanceSpec.value = JSON.stringify(fields, null, ' ');
      formDetailsElement.appendChild(instanceSpec);

      var submit = document.createElement("button");
      submit.innerHTML = "Submit";
      //submit.setAttribute("onclick","formHandler()");
      formDetailsElement.appendChild(submit);

    }
  }
  xhttp.send();
}


function create_resource_prev(resource) {

  document.getElementById("input-form").style.display = "block";

  elementsToHide = ["metrics-details","man-page","consumption_string_id","num_of_instances","create-status"];
  hideElements(elementsToHide);

  var fields
  url = "/service/" + resource + "/field_names";
  var xhr = new XMLHttpRequest();
  xhr.open('GET', url, true);
  xhr.onreadystatechange = function () {
    if (this.readyState == 4 && this.status == 200) {
        fieldData = this.responseText;
        console.log(fieldData);
        console.log("-------");
        data1 = JSON.parse(fieldData)
        fields = data1["fields"]

        formDetailsElement = document.getElementById("form-details");
        header = document.createElement("h4");
        header.innerHTML = "Enter data";
        formDetailsElement.appendChild(header);

        var ul = document.createElement("ul");
        ul.setAttribute("list-style","none");
        ul.setAttribute("padding-left",0);
        formDetailsElement.appendChild(ul);

        for (const val of fields) {
            var br = document.createElement("br");
            var div = document.createElement("div");

            var label = document.createElement("label");
            label.setAttribute("for",val);
            var labelName = document.createElement("b");
            labelName.innerHTML = val;
            label.appendChild(labelName);
            formDetailsElement.appendChild(label);

            formDetailsElement.appendChild(div);

            var input = document.createElement("input");
            input.setAttribute("type", "text");
            input.setAttribute("name",val);
            formDetailsElement.appendChild(input);
            formDetailsElement.appendChild(br);
      }

      var submit = document.createElement("button");
      submit.innerHTML = "Submit";
      //submit.setAttribute("onclick","formHandler()");
      formDetailsElement.appendChild(submit);
    }
  };
  xhr.send();
}

// Not used
function formHandler() {
  form = document.getElementById("form-details");
  url = form.getAttribute('action');

  currentURL = window.location.href;

  var data = new FormData(form);

  var xhr = new XMLHttpRequest();
  xhr.open('POST', url, true);
  xhr.onreadystatechange = function() {
    //window.location.href = currentURL;
    console.log(this.responseText);
    //alert(this.responseText);
    /*var create_status = document.getElementById("create-status");
    create_status.setAttribute("style","display:block;")
    create_status.innerHTML = this.responseText*/
  };
  xhr.onerror = function() {
    //window.location.href = currentURL;
    console.log(this.responseText);
    //alert(this.responseText);
    //var create_status = document.getElementById("create-status");
    //create_status.setAttribute("style","display:block;")
    //create_status.innerHTML = this.responseText
  }
  xhr.send(data);
}

function hideElements(fields) {
  for (const val of fields) {
    document.getElementById(val).style.display = "none";
  }
}
