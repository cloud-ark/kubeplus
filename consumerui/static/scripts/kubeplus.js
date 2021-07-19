function validateData() {
  return true;
}

function get_all_resources(resource){
  //alert("Hello World");
  displayString = "1. abc-org-tenant1 <br> 2. abc-org-tenant2"

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

  elementsToHide = ["input-form", "man-page", "num_of_instances","create-status"];
  hideElements(elementsToHide);

  document.getElementById("metrics-details").style.display = "block";
  document.getElementById("metrics-details").setAttribute("class","table table-condensed table-striped table-bordered");
  document.getElementById("metrics-details").setAttribute("width","100%");

  myArr = res_string.split(",")

  console.log("Name:" + myArr[0]);
  resName = myArr[0];
  namespace = myArr[1];
  service = myArr[2];
  element = document.getElementById("consumption_string_id");
  element.innerHTML = "Consumption metrics for " + resName;
  element.style.display = "block";

  // TODO: Make Ajax call to get metrics for <resName, namespace, service>
  cpu = 23;
  memory = 50;
  storage = 100;
  nw_ingress = 888888;
  nw_egress = 444444;

  element = document.getElementById("total_cpu");
  element.innerHTML = cpu + "<br> (millicores)";

  element = document.getElementById("total_memory");
  element.innerHTML = memory + "<br> (mebibytes)" ;

  element = document.getElementById("total_storage");
  element.innerHTML = storage + "<br> (Giga bytes)" ;

  element = document.getElementById("total_nw_ingress");
  element.innerHTML = nw_ingress + "<br> (bytes)" ;

  element = document.getElementById("total_nw_egress");
  element.innerHTML = nw_egress + "<br> (bytes)" ;
}

function create_resource(resource) {

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