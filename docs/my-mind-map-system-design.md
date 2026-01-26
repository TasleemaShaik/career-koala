# My Mind Map

## First things first, database design:
- I need a postgreSQL database with a local running one as of now(I can connect cloud postgres from GCP in the future).
- It should contain 4 major tables - Job Applications, Coding, Projects, Networking.
- Job Applications should have unique id to track the count, jobTitle, jobLink, appliedDate, resultDate, etc 
- Coding should majorly have leetcode style problem details, unique id to track the count, leetCodeProblemNumber, Pattern it falls into, linkTotheProblem, alreadySolved or not boolean, etc 
- Projects should have unique id, project name, github repo link, actively doing that project(boolean), techstack used(list[string]) 
- Networking Primarily should have a unique id, person name, how I met them, connected in LinkedIn(boolean), their current company, current Position, etc 
- Daily goals - unique id, description of the goal, foreign key connection with one of the 4 tables above based on the type of goal, take their primary key as foreign as reference here to track. For Example, you need to solve this leetcode problem today, did you do it?, completed(boolean) 
- Weekly Goals - similar to weekly goals
- Monthly Goals - Similar to daily and weekly goals
- Meetings - THis is majorly for tracking coffee chats, meetings, interviews, networking sessions, some events that could help for job search(job prep, alumni talks etc), unique id, session name, session type(virtual, in-person), session location details(zoom link or address), company/org who is organising this etc 

## Secondly, what our GO ADK should do
- Multi-Agent Architecture: We should have 4 agents to track these 4 different aspects(Job Applications, Coding, Projects, Networking) - Main Agent would handle everything based on the details whatever user is saying, give 4 options to select 4 agents, 5 as Other to talk about random things which main agent would do a google search and respond
- All Agents should have a input component giving user a scope to insert/update related data into the application
- First Job Applications Agent, once it gets data about which company we are appplying, it should do an internal google research about the company any blogs, how their interview process is, are there any executives that could help with tips as an insider who prefers actively posting on social media so that we know they are active and could help us.
- Secondly, Coding Agent, user could provide a single problem at any point of time, also a list of CSV for a targeted company, parse & store all those in the DB directly to track in the goals
- Thirdly, Projects, as we take the existing project data, do an analysis(use google search or gen AI models) based on the resume what they are expert at and what could help them more new project ideas, what technologies they need to improve. For example, they mentioned apache KAFKA in their resume they do not have a single project based on that. In the next project ideas, you could suggest that!
- Finally, networking once you get all the seed data from the user. Make sure you process all the data and run some ML algorithm to see potential job applications in those companies and compare if the user already did apply to those companies, if the result was already declared didn't go through, recommending do it again if there are no restrictions to do applications.

## UI Component
- I'm thinking basic next.js app would be good.
- In the main page, Four slots/squares divided each for every agent's input at any point of time, right side of the corner would have some notification bar about daily/weekly/monthly goals if at that point of time no entries in the goals no notifications for that.
- Bottom part of the page would have chatbox to talk to the agent randomly about anything which internally does google search and respond.

## Deployment
- As of now, I'm not sure, let's see. For now, we will just execute in the local and go for GCP and Kaggle part later