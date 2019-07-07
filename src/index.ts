import * as Discord from 'discord.js';
import * as dotenv from 'dotenv';

const client = new Discord.Client();
dotenv.config();

client.on('ready', () => {
  console.log('Bot connected');
});

client.on('message', message => {
  if (message.content.startsWith('!test')) {
    message.channel.send("I'm Alive :D");
  }
});

client.login(process.env.API_KEY); // haha get pranked API key scraper!
