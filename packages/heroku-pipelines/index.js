exports.topic = {
  name: 'pipelines',
  // this is the help text that shows up under `heroku help`
  description: 'manage collections of apps in pipelines'
};

exports.commands = [
  require('./commands/pipelines/create'),
  require('./commands/pipelines/list'),
  require('./commands/pipelines/show')
];
