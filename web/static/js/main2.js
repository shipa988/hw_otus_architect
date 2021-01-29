var lastid;
(function() {
    var lazyLoad = function(isfriend){
        var div = document.getElementById('contDiv'), i=0, j=0;
        div.onscroll = function(){
            if((div.scrollHeight-div.offsetHeight) <= div.scrollTop+10){
                function animFn(){
                    i++;
                    if(i%7==1){
                        j++;
                        if(j%5==1){
                            loadfriends(isfriend)
                            /*                            var cars = ["BMW", "Volvo", "Saab", "Ford", "Fiat", "Audi"];
                                                        var T;
                                                        for (T = 0; T < cars.length; T++) {
                                                            div.children[0].insertBefore(div.children[0].lastElementChild.cloneNode(false), div.children[0].nextElementSibling);
                                                            div.children[0].lastElementChild.innerHTML ="<img src='https://i.imgur.com/nkN3Mv0.jpg' /><div class='name'>Sally Lin</div>";
                                                        }*/

                        }
                        else{
                            setTimeout(animFn, 0);
                        }
                    }
                    else{
                        setTimeout(animFn, 0);
                    }
                }
                setTimeout(animFn, 0);
            }
        }
    }
    lazyLoad();
}());

function loadfriends(isfriends){
    var xhr = new XMLHttpRequest();
    xhr.withCredentials = true;

    xhr.addEventListener("readystatechange", function () {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                var jsonData = JSON.parse(xhr.responseText);
                var div = document.getElementById('contDiv'), i=0, j=0;
                for (var i = 0; i < jsonData.friends; i++) {
                    div.children[0].insertBefore(div.children[0].lastElementChild.cloneNode(false), div.children[0].nextElementSibling);
                    div.children[0].lastElementChild.innerHTML ="<img src='https://i.imgur.com/nkN3Mv0.jpg' /><div class='name'>jsonData.friends[i].Name+' '+jsonData.friends[i].Surname</div>";
                }
                lastid=jsonData.friends[jsonData.friends.length-1]
            }
        }
    });
    if (isfriends) {
        xhr.open("GET", "/getpeople?friends=1&lastid="+lastid+"&limit=10");
    } else {
        xhr.open("GET", "/getpeople?friends=0&lastid="+lastid+"&limit=10");
    }
    xhr.send();
}