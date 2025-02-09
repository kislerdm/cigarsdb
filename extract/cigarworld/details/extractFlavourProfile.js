function readCommunityFlavourProfile() {
    const d = document.getElementsByClassName("aromaimg")[0];
    const flavours = d.attributes['data-rub'].nodeValue == "t" ? NameArrObj.AromaTabacNamenArr : NameArrObj.AromaNamenArr;
    const rawData = d.attributes['data-content'].nodeValue;
    const o = {};
    let sum = 0;
    for (let i = 0; i < rawData.length; i++) {
        const v = parseInt(rawData[i]);
        o[flavours[i]] = v;
        sum += v;
    }
    for (let i = 0; i < rawData.length; i++) {
        o[flavours[i]] = o[flavours[i]] / sum;
    }
    return o;
}