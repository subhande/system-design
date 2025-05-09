# Low Level Design - Stackoverflow

"""
## Requirements
1. Users can post questions, answer questions, and comment on questions and answers.
2. Users can vote on questions and answers.
3. Questions should have tags associated with them.
4. Users can search for questions based on keywords, tags, or user profiles.
5. The system should assign reputation score to users based on their activity and the quality of their contributions.
6. The system should handle concurrent access and ensure data consistency.

# Actors
## User
- Register
- Login and Logout
- Post Questions
- Post Answers
- Comment on Questions and Answers
- Vote on Questions and Answers
- Add tags on posted questions
- Search keywords, tags, or user profiles

## System
- Log User Activity
- Reputation Score
- Handle Concurrent Access

"""
from datetime import datetime
from abc import ABC, abstractmethod
from typing import Dict


class User:
    def __init__(self, user_id: int, username: str, email: str):
        self.user_id = user_id
        self.username = username
        self.email = email
        self.reputation = 0
        self.questions = []
        self.answers = []
        self.comments = []

    def ask_question(self, title: str, content: str, tags: list):
        question = Question(self, title, content, tags)
        self.questions.append(question)
        self.update_reputation(5)  # Gain 5 reputation for asking a question
        return question

    def answer_question(self, question, content: str):
        answer = Answer(self, question, content)
        self.answers.append(answer)
        question.add_answer(answer)
        self.update_reputation(10)
        return answer

    def comment_on(self, commentable, content: str):
        comment = Comment(self, content)
        self.comments.append(comment)
        commentable.add_comment(comment)
        self.update_reputation(2)
        return comment

    def update_reputation(self, value: int):
        self.reputation += value
        if self.reputation < 0:
            self.reputation = 0

class Votable(ABC):
    @abstractmethod
    def vote(self, user: User, value: int):
        pass

    @abstractmethod
    def get_vote_count(self) -> int:
        pass


class Commentable(ABC):
    @abstractmethod
    def add_comment(self, comment):
        pass

    @abstractmethod
    def get_comments(self) -> list:
        pass


class Vote:
    def __init__(self, user: User, value: int):
        self.user = user
        self.value = value
        self.date = datetime.now()


class Question(Votable, Commentable):
    def __init__(self, author: User, title: str, content: str, tags: list):
        self.id = id(self)
        self.author = author
        self.title = title
        self.content = content
        self.tags = [Tag(name) for name in tags]
        self.creation_date = datetime.now()
        self.answers = []
        self.votes = []
        self.comments = []

    def add_answer(self, answer):
        if answer not in self.answers:
            self.answers.append(answer)

    def add_comment(self, comment):
        self.comments.append(comment)

    def get_comments(self) -> list:
        return self.comments.copy()

    def vote(self, user: User, value: int):
        if value not in [-1, 1]:
            raise ValueError("Vote value must be either 1 or -1")
        self.votes = [v for v in self.votes if v.user != user]
        self.votes.append(Vote(user, value))
        user.update_reputation(value * 5)  # +5 for upvote, -5 for downvote

    def get_vote_count(self) -> int:
        return sum(v.value for v in self.votes)



class Answer(Votable, Commentable):
    def __init__(self, author: User, question: Question, content: str):
        self.id = id(self)
        self.author = author
        self.question = question
        self.content = content
        self.creation_date = datetime.now()
        self.votes = []
        self.comments = []
        self.is_accepted = False

    def add_comment(self, comment):
        self.comments.append(comment)

    def get_comments(self) -> list:
        return self.comments.copy()

    def vote(self, user: User, value: int):
        if value not in [-1, 1]:
            raise ValueError("Vote value must be either 1 or -1")
        self.votes = [v for v in self.votes if v.user != user]
        self.votes.append(Vote(user, value))
        user.update_reputation(value * 10)  # +10 for upvote, -10 for downvote

    def get_vote_count(self) -> int:
        return sum(v.value for v in self.votes)

    def accept(self):
        if self.is_accepted:
            raise ValueError("This answer is already accepted")
        self.is_accepted = True
        self.author.update_reputation(15)  # +15 reputation for accepted answer

class Comment:
    def __init__(self, author: User, content: str):
        self.id = id(self)
        self.author = author
        self.content = content
        self.creation_date = datetime.now()


class Tag:
    def __init__(self, name: str):
        self.id = id(self)
        self.name = name


# Main class to manage the StackOverflow-like system

class StackOverflow:
    def __init__(self):
        self.users: Dict[int, User] = {}
        self.questions: Dict[int, Question] = {}
        self.answers: Dict[int, Answer] = {}
        self.tags : Dict[str, Tag] = {}

    def create_user(self, username: str, email: str) -> User:
        user_id = len(self.users) + 1
        user = User(user_id, username, email)
        self.users[user_id] = user
        return user

    def ask_question(self, user: User, title: str, content: str, tags: list) -> Question:
        question = user.ask_question(title, content, tags)
        self.questions[question.id] = question
        for tag in question.tags:
            self.tags.setdefault(tag.name, tag)
        return question

    def answer_question(self, user: User, question: Question, content: str) -> Answer:
        answer = user.answer_question(question, content)
        self.answers[answer.id] = answer
        return answer

    def add_comment(self, user: User, commentable, content: str):
        return user.comment_on(commentable, content)

    def vote_question(self, user: User, question: Question, value: int):
        question.vote(user, value)

    def vote_answer(self, user: User, answer: Answer, value: int):
        answer.vote(user, value)

    def accept_answer(self, answer: Answer):
        answer.accept()

    def search_questions(self, query: str) -> list:
        return [q for q in self.questions.values() if
                query.lower() in q.title.lower() or
                query.lower() in q.content.lower() or
                any(query.lower() == tag.name.lower() for tag in q.tags)]


    def get_questions_by_user(self, user):
        return user.questions

    def get_user(self, user_id):
        return self.users.get(user_id)

    def get_question(self, question_id):
        return self.questions.get(question_id)

    def get_answer(self, answer_id):
        return self.answers.get(answer_id)

    def get_tag(self, name: str):
        return self.tags.get(name)



class StackOverflowDemo:
    @staticmethod
    def run():
        system = StackOverflow()

        # Create users
        alice = system.create_user("Alice", "alice@example.com")
        bob = system.create_user("Bob", "bob@example.com")
        charlie = system.create_user("Charlie", "charlie@example.com")

        # Alice asks a question
        java_question = system.ask_question(alice, "What is polymorphism in Java?",
                                            "Can someone explain polymorphism in Java with an example?",
                                            ["java", "oop"])

        # Bob answers Alice's question
        bob_answer = system.answer_question(bob, java_question,
                                            "Polymorphism in Java is the ability of an object to take on many forms...")

        # Charlie comments on the question
        system.add_comment(charlie, java_question, "Great question! I'm also interested in learning about this.")

        # Alice comments on Bob's answer
        system.add_comment(alice, bob_answer, "Thanks for the explanation! Could you provide a code example?")

        # Charlie votes on the question and answer
        system.vote_question(charlie, java_question, 1)  # Upvote
        system.vote_answer(charlie, bob_answer, 1)  # Upvote

        # Alice accepts Bob's answer
        system.accept_answer(bob_answer)

        # Bob asks another question
        python_question = system.ask_question(bob, "How to use list comprehensions in Python?",
                                            "I'm new to Python and I've heard about list comprehensions. Can someone explain how to use them?",
                                            ["python", "list-comprehension"])

        # Alice answers Bob's question
        alice_answer = system.answer_question(alice, python_question,
                                            "List comprehensions in Python provide a concise way to create lists...")

        # Charlie votes on Bob's question and Alice's answer
        system.vote_question(charlie, python_question, 1)  # Upvote
        system.vote_answer(charlie, alice_answer, 1)  # Upvote

        # Print out the current state
        print(f"Question: {java_question.title}")
        print(f"Asked by: {java_question.author.username}")
        print(f"Tags: {', '.join(tag.name for tag in java_question.tags)}")
        print(f"Votes: {java_question.get_vote_count()}")
        print(f"Comments: {len(java_question.get_comments())}")
        print(f"\nAnswer by {bob_answer.author.username}:")
        print(bob_answer.content)
        print(f"Votes: {bob_answer.get_vote_count()}")
        print(f"Accepted: {bob_answer.is_accepted}")
        print(f"Comments: {len(bob_answer.get_comments())}")

        print("\nUser Reputations:")
        print(f"Alice: {alice.reputation}")
        print(f"Bob: {bob.reputation}")
        print(f"Charlie: {charlie.reputation}")

        # Demonstrate search functionality
        print("\nSearch Results for 'java':")
        search_results = system.search_questions("java")
        for q in search_results:
            print(q.title)

        print("\nSearch Results for 'python':")
        search_results = system.search_questions("python")
        for q in search_results:
            print(q.title)

        # Demonstrate getting questions by user
        print("\nBob's Questions:")
        bob_questions = system.get_questions_by_user(bob)
        for q in bob_questions:
            print(q.title)

if __name__ == "__main__":
    StackOverflowDemo.run()
