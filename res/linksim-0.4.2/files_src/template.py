defualt_template = """

<form action="{{ action }}" method="post" enctype="multipart/form-data">
{% for k,form in forms.items() %}
<div class="form-group">
        {% if form["type"] == "text" %}
        <div class="form-inline"  style="margin:10px;" >
            <label class="form-label" style="margin-right: 10px; font-weight: 600;">{{ form['name']}}: </label>    
            <input type="text" class="form-control" placeholder="{{ form['name'] }}" name="{{ form['name'] }}" id="id-{{ form['name'] }}">
        {% elif form['type'] == 'num' %}
        <div class="form-inline"  style="margin:10px;" >
            <label class="form-label" style="margin-right: 10px; font-weight: 600;">{{ form['name']}}: </label>    
            <input type="number" class="form-control" placeholder="{{ form['name'] }}" name="{{ form['name'] }}" id="id-{{ form['name'] }}">
        {% elif form['type'] == 'img' %}
        <div class="form-inline"  style="margin:10px;" >
            <label class="form-label" style="margin-right: 10px; font-weight: 600;">{{ form['name']}}: </label>    
            <input type="file" multiple  name="{{ form['name']}}"  alt="img">
        {% else %}
        <div class="form-group"  style="margin:10px;" >
            <label class="form-label" style="margin-right: 10px; font-weight: 600;">{{ form['name']}}: </label>    
        <textarea class="form-control" name="{{ form['name'] }}"placeholder="{{ form['name'] }}" id="id-{{ form['name'] }}" cols="30" rows="10"></textarea>
        {% endif %}
    </div>
</div>

{% endfor %}
{% for k, v in hiddens.items() %}
    <input type="hidden" name="{{k}}" value="{{v}}">
{% endfor %}
<script>
    function isEmpty(val) {
        if (typeof(val) == 'number') {
            val += '';
        }
        var str = val || '';
        return $.trim(str).length == 0;
    }
    function checkEmpty(){
        ins = $("input")

        for(var j=0;j< ins.length;j++){
            if (isEmpty($(ins[j]).val()) == true){
                alert("不能为空!!!填不出就写无")
                return false
            }
        }
        ins = $("textarea")

        for(var j=0;j< ins.length;j++){
            if (isEmpty($(ins[j]).val()) == true){
                alert("不能为空!!!填不出就写无")
                return false
            }
        }
        return true;
    }
</script>
<div class="form-group"><input type="submit" class="form-control" value="提交" onclick="return checkEmpty()"></div>
</form>
"""