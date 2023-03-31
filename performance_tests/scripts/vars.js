import { group } from "k6";
export default function() {
    // Declare variables here for global use
    let var1;
    group("function_1", function() {var1 = doPage1()});
    group("function2", function(){doPage2(var1)});
};

function doPage1 () {
    return "something";
};

function doPage2 (var1) {
    const url = "url2.com";
    console.log(JSON.stringify({ "token": `${var1}`}));
};