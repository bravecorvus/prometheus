var vm = new Vue({
  el: '#app',
  data: {
    alarms: [],
    trackinfo: []
  },
  created() {
    $.getJSON('../json/alarms.json')
      .done(data => {this.alarms = data;});
  }
})
