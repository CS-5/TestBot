import * as Discord from 'discord.js';

const client = new Discord.Client();

client.on('ready', () => {
  console.log('Bot connected');
});

client.on('message', message => {
  if (message.content.startsWith('!test')) {
    message.channel.send("I'm Alive :D");
  }
});

client.login('NTk1OTk4ODI5MzEzMjYxNTY4.XR33Eg.uVCdZUnTh5rUEN3i4NE-bEj0nqI');
