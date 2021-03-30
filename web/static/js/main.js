var lastid = 0;

function lazyLoad(div, isfriend, name, surname) {
    if ((div.scrollHeight - div.offsetHeight) <= div.scrollTop) {
        var i = 0, j = 0;

        function animFn() {
            i++;
            if (i % 5 == 1) {
                j++;
                if (j % 7 == 1) {
                    loadfriends(isfriend, name, surname)
                } else {
                    setTimeout(animFn, 5);
                }
            } else {
                setTimeout(animFn, 5);
            }
        }

        setTimeout(animFn, 5);
    }
}

function loadnews() {
    var xhr = new XMLHttpRequest();
    xhr.withCredentials = true;

    xhr.addEventListener("readystatechange", function () {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                var jsonData = JSON.parse(xhr.responseText);
                var div = document.getElementById('contDiv');
                for (var i = 0; i < jsonData.length; i++) {
                    if (div.children[0].children.length == 0) {
                        var ul = document.getElementById("news-list");
                        var li = document.createElement("li");
                        li.classList.add('news');
                        li.setAttribute("onClick", "location.href='/profile?id=" + jsonData[i].AuthorId + "'")
                        var d = new Date(jsonData[i].Time);
                        switch (jsonData[i].AuthorGen) {
                            case "Male":
                                li.innerHTML = "<img src='static/img/male.jpg' style='width:90px'/><span>" + jsonData[i].AuthorName + " " + jsonData[i].AuthorSurName
                                    + "</span><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";
                                break;
                            case "Female":
                                li.innerHTML = "<img src='static/img/female.jpg' style='width:90px'/><span>" + jsonData[i].AuthorName + " " + jsonData[i].AuthorSurName
                                    + "</span><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";
                                break;
                            default:
                                li.innerHTML = "<img src='static/img/other.png' style='width:90px'/><span>" + jsonData[i].AuthorName + " " + jsonData[i].AuthorSurName
                                    + "</span><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";

                        }
                        ul.appendChild(li)
                        continue
                    }
                    div.children[0].insertBefore(div.children[0].lastElementChild.cloneNode(false), div.children[0].nextElementSibling);
                    div.children[0].lastElementChild.setAttribute("onClick", "location.href='/profile?id=" + jsonData[i].AuthorId + "'")
                    var d = new Date(jsonData[i].Time);
                    switch (jsonData[i].AuthorGen) {
                        case "Male":
                            div.children[0].lastElementChild.innerHTML = "<img src='static/img/male.jpg' style='width:90px'/><span>" + jsonData[i].AuthorName + " " + jsonData[i].AuthorSurName
                                + "</span><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";
                            break;
                        case "Female":
                            div.children[0].lastElementChild.innerHTML = "<img src='static/img/female.jpg' style='width:90px'/><span>" + jsonData[i].AuthorName + " " + jsonData[i].AuthorSurName
                                + "</span><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";
                            break;
                        default:
                            div.children[0].lastElementChild.innerHTML = "<img src='static/img/other.png' style='width:90px'/><span>" + jsonData[i].AuthorName + " " + jsonData[i].AuthorSurName
                                + "</span><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";
                    }

                }
                lastid = jsonData[jsonData.length - 1].Id
            }
        }
    });
    xhr.open("GET", "/getnews");

    xhr.send();
}

function loadmynews() {
    var xhr = new XMLHttpRequest();
    xhr.withCredentials = true;

    xhr.addEventListener("readystatechange", function () {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                var jsonData = JSON.parse(xhr.responseText);
                var div = document.getElementById('contDiv');
                for (var i = 0; i < jsonData.length; i++) {
                    if (div.children[0].children.length == 0) {
                        var ul = document.getElementById("news-list");
                        var li = document.createElement("li");
                        li.classList.add('news');
                        var d = new Date(jsonData[i].Time);
                        switch (jsonData[i].AuthorGen) {
                            case "Male":
                                li.innerHTML = "<img src='static/img/male.jpg' style='width:90px'/><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";
                                break;
                            case "Female":
                                li.innerHTML = "<img src='static/img/female.jpg' style='width:90px'/><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";
                                break;
                            default:
                                li.innerHTML = "<img src='static/img/other.png' style='width:90px'/><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";

                        }
                        ul.appendChild(li)
                        continue
                    }
                    div.children[0].insertBefore(div.children[0].lastElementChild.cloneNode(false), div.children[0].nextElementSibling);
                    var d = new Date(jsonData[i].Time);
                    switch (jsonData[i].AuthorGen) {
                        case "Male":
                            div.children[0].lastElementChild.innerHTML = "<img src='static/img/male.jpg' style='width:90px'/><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";
                            break;
                        case "Female":
                            div.children[0].lastElementChild.innerHTML = "<img src='static/img/female.jpg' style='width:90px'/><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";
                            break;
                        default:
                            div.children[0].lastElementChild.innerHTML = "<img src='static/img/other.png' style='width:90px'/><p><span>" + jsonData[i].Title + "</span></p><p>" + jsonData[i].Text + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";
                    }
                }
                lastid = jsonData[jsonData.length - 1].Id
            }
        }
    });
    xhr.open("GET", "/getmynews");

    xhr.send();
}

// Create a new list item when clicking on the "Add" button
function newElement() {
    var inputValue = document.getElementById("title").value;
    var inputContent = document.getElementById("content").value;

    var xhr = new XMLHttpRequest();

    var body = 'title=' + encodeURIComponent(inputValue) +
        '&content=' + encodeURIComponent(inputContent);

    xhr.open("POST", '/postnews', true);
    xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');

    xhr.onreadystatechange = function () {
        if (this.readyState != 4) return;
        var ul = document.getElementById("news-list");
        var li = document.createElement("li");
        li.classList.add('news');
        var d = new Date();
        li.innerHTML = "<img src='static/img/other.png' style='width:90px'/><p><span>" + inputValue + "</span></p><p>" + inputContent + "</p><div class='time-right'>" + d.toLocaleString('en-GB') + "</div>";

        if (inputValue === '') {
            alert("You must write something!");
        } else {
            document.getElementById("news-list").appendChild(li);
        }
        document.getElementById("title").value = "";
        document.getElementById("content").value = "";
        var span = document.createElement("SPAN");
        ul.prepend(li);
    }
    xhr.send(body);
}

function loadfriends(isfriends, name, surname) {
    var xhr = new XMLHttpRequest();
    xhr.withCredentials = true;

    xhr.addEventListener("readystatechange", function () {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                var jsonData = JSON.parse(xhr.responseText);
                var div = document.getElementById('contDiv');
                for (var i = 0; i < jsonData.length; i++) {
                    if (div.children[0].children.length == 0) {
                        var ul = document.getElementById("friend-list");
                        var li = document.createElement("li");
                        li.classList.add('friend');
                        li.setAttribute("onClick", "location.href='/profile?id=" + jsonData[i].Id + "'")
                        switch (jsonData[i].Gen) {
                            case "Male":
                                li.innerHTML = "<img src='static/img/male.jpg' /><div class='name'>" + jsonData[i].Name + " " + jsonData[i].SurName + "</div>";
                                break;
                            case "Female":
                                li.innerHTML = "<img src='static/img/female.jpg' /><div class='name'>" + jsonData[i].Name + " " + jsonData[i].SurName + "</div>";
                                break;
                            default:
                                li.innerHTML = "<img src='static/img/other.png' /><div class='name'>" + jsonData[i].Name + " " + jsonData[i].SurName + "</div>";

                        }
                        ul.appendChild(li)
                        continue
                    }
                    div.children[0].insertBefore(div.children[0].lastElementChild.cloneNode(false), div.children[0].nextElementSibling);
                    div.children[0].lastElementChild.setAttribute("onClick", "location.href='/profile?id=" + jsonData[i].Id + "'")
                    switch (jsonData[i].Gen) {
                        case "Male":
                            div.children[0].lastElementChild.innerHTML = "<img src='static/img/male.jpg' /><div class='name'>" + jsonData[i].Name + " " + jsonData[i].SurName + "</div>";
                            break;
                        case "Female":
                            div.children[0].lastElementChild.innerHTML = "<img src='static/img/female.jpg' /><div class='name'>" + jsonData[i].Name + " " + jsonData[i].SurName + "</div>";
                            break;
                        default:
                            div.children[0].lastElementChild.innerHTML = "<img src='static/img/other.png' /><div class='name'>" + jsonData[i].Name + " " + jsonData[i].SurName + "</div>";

                    }

                }
                lastid = jsonData[jsonData.length - 1].Id
            }
        }
    });
    if (isfriends) {
        xhr.open("GET", "/getpeople?friends=1&lastid=" + lastid + "&limit=10");
    } else {
        xhr.open("GET", "/getpeople?friends=0&lastid=" + lastid + "&limit=10&name=" + name + "&surname=" + surname);
    }
    xhr.send();
}
